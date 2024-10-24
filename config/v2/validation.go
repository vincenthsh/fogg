package v2

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/chanzuckerberg/fogg/errs"
	"github.com/chanzuckerberg/fogg/util"
	multierror "github.com/hashicorp/go-multierror"
	goVersion "github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	validator "gopkg.in/go-playground/validator.v9"
)

const rootPath = "terraform"

var validCICommands = map[string]struct{}{
	"check": {},
	"lint":  {},
}

// Validate validates the config
func (c *Config) Validate(fs afero.Fs) ([]string, error) {
	if c == nil {
		return nil, errs.NewInternal("config is nil")
	}

	var errs *multierror.Error
	var warnings []string

	v := validator.New()
	// This func gives us the ability to get the full path for a field deeply
	// nested in the structure
	// https://github.com/go-playground/validator/issues/323#issuecomment-343670840
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("yaml"), ",", 2)[0]

		if name == "-" {
			return ""
		}
		return name
	})

	errs = multierror.Append(errs, v.Struct(c))
	errs = multierror.Append(errs, c.validateExtraVars())
	errs = multierror.Append(errs, c.validateInheritedStringField("owner", OwnerGetter, nonEmptyString))
	errs = multierror.Append(errs, c.validateInheritedStringField("project", ProjectGetter, nonEmptyString))
	errs = multierror.Append(errs, c.validateInheritedStringField("terraform version", TerraformVersionGetter, nonEmptyString))
	errs = multierror.Append(errs, c.ValidateBackends())
	errs = multierror.Append(errs, c.validateAWSProviderAuth())
	errs = multierror.Append(errs, c.ValidateAWSProviders())
	errs = multierror.Append(errs, c.ValidateSnowflakeProviders())
	errs = multierror.Append(errs, c.ValidateBlessProviders())
	errs = multierror.Append(errs, c.ValidateOktaProviders())
	errs = multierror.Append(errs, c.ValidateGenericProviders())
	errs = multierror.Append(errs, c.validateModules())
	errs = multierror.Append(errs, c.ValidateTravis())
	errs = multierror.Append(errs, c.ValidateGithubActionsCI())
	errs = multierror.Append(errs, c.validateTFE())
	errs = multierror.Append(errs, c.ValidateFileDependencies(fs))

	// refactor to make it easier to manage these
	w, e := c.ValidateToolsTfLint()
	warnings = append(warnings, w...)
	errs = multierror.Append(errs, e)

	return warnings, errs.ErrorOrNil()
}

func ValidateAWSProvider(p *AWSProvider, component string) error {
	var errs *multierror.Error
	if p == nil {
		return nil // nothing to validate
	}

	if p.Region == nil {
		errs = multierror.Append(errs, fmt.Errorf("aws provider region for %s", component))
	}

	if (p.Profile == nil) == (p.Role == nil) {
		errs = multierror.Append(errs, fmt.Errorf("aws provider must have exactly one of profile or role %s ", component))
	}

	if p.Version == nil {
		errs = multierror.Append(errs, fmt.Errorf("aws provider version for %s ", component))
	}

	if p.AccountID == nil || *p.AccountID == "" {
		errs = multierror.Append(errs, fmt.Errorf("aws provider account id for %s", component))
	}
	return errs.ErrorOrNil()
}

// ValidateBackend will check the resolved configuration for a backend and return any errors it
// finds
func ValidateBackend(backend *Backend, component string) error {
	if backend == nil {
		return nil
	}

	var errs *multierror.Error

	if backend.Kind == nil {
		errs = multierror.Append(errs, fmt.Errorf("unable to resolve backend for component %s", component))

		return errs.ErrorOrNil()
	}

	if *backend.Kind == "s3" {
		if backend.Bucket == nil {
			errs = multierror.Append(errs, fmt.Errorf("when backend kind == 's3', bucket is required (component %s)", component))
		}

		if backend.Region == nil {
			errs = multierror.Append(errs, fmt.Errorf("when backend kind == 's3', region is required (component %s)", component))
		}

		if (backend.Profile != nil && backend.Role != nil) ||
			(backend.Profile == nil && backend.Role == nil) {
			errs = multierror.Append(errs, fmt.Errorf("when backend kind == 's3', exactly one of profile or role must be set (component %s)", component))
		}

		if backend.Role != nil && backend.AccountID == nil {
			errs = multierror.Append(errs, fmt.Errorf("when backend kind == 's3' and role is set, account_id must also be set (component %s)", component))
		}

		return errs.ErrorOrNil()
	}

	if *backend.Kind == "remote" {
		if backend.HostName == nil {
			errs = multierror.Append(errs, fmt.Errorf("when backend kind == 'remote', host_name is required (component %s)", component))
		}

		if backend.Organization == nil {
			errs = multierror.Append(errs, fmt.Errorf("when backend kind == 'remote', organization is required (component %s)", component))
		}

		return errs.ErrorOrNil()
	}

	errs = multierror.Append(errs, fmt.Errorf("invalid backend kind %#v for component %s", backend.Kind, component))

	return errs.ErrorOrNil()
}

func (c *Config) ValidateAWSProviders() error {
	var errs *multierror.Error

	c.WalkComponents(func(component string, comms ...Common) {
		v := ResolveAWSProvider(comms...)
		if e := ValidateAWSProvider(v, component); e != nil {
			errs = multierror.Append(errs, e)
		}
	})

	return errs.ErrorOrNil()
}

func (c *Config) ValidateBackends() error {
	var errs *multierror.Error

	c.WalkComponents(func(component string, comms ...Common) {
		// NOTE[JH]: don't require the global/default have a backend
		// Some repos manage several TFE organizations and requiring
		// a global backend will inject remote states for which they cannot
		// authenticate to.
		if component == "global" {
			return
		}
		backendConfig := ResolveBackend(comms...)

		if e := ValidateBackend(backendConfig, component); e != nil {
			errs = multierror.Append(errs, e)
		}
	})
	return errs.ErrorOrNil()
}

func (p *BlessProvider) Validate(component string) error {
	var errs *multierror.Error
	if p == nil {
		return nil // nothing to do
	}

	if p.AWSProfile == nil && p.RoleArn == nil {
		errs = multierror.Append(errs, fmt.Errorf("bless provider requires aws_profile or role_arn in %s", component))
	}
	if p.AWSRegion == nil {
		errs = multierror.Append(errs, fmt.Errorf("bless provider aws_region required in %s", component))
	}
	return errs
}

func (p *GenericProvider) Validate(name string, component string) error {
	var errs *multierror.Error
	if p == nil {
		return nil // nothing to do
	}

	if len(p.Source) == 0 {
		errs = multierror.Append(errs, fmt.Errorf("required provider %q requires non-empty source in %s", name, component))
	}
	if p.Version == nil || len(*p.Version) == 0 {
		errs = multierror.Append(errs, fmt.Errorf("required provider %q requires version in %s", name, component))
	}
	if p.Config["assume_role"] != nil {
		// "role" key must be provided in assume_role map
		valueMap, ok := p.Config["assume_role"].(map[string]any)
		if !ok {
			errs = multierror.Append(errs, fmt.Errorf("required provider %q requires 'assume_role' to be map (cast failed) in %s", name, component))
		} else if valueMap["role"] == nil && valueMap["role_arn"] == nil {
			errs = multierror.Append(errs, fmt.Errorf("required provider %q requires 'assume_role' map to contain at least 'role' key in %s (received: %v)", name, component, valueMap))
		}
	}
	return errs.ErrorOrNil()
}

func (p *SnowflakeProvider) Validate(component string) error {
	var errs *multierror.Error
	if p == nil {
		return nil // nothing to do
	}

	if p.Account == nil {
		errs = multierror.Append(errs, fmt.Errorf("snowflake provider account must be set in %s", component))
	}

	if p.Role == nil {
		errs = multierror.Append(errs, fmt.Errorf("snowflake provider role must be set in %s", component))
	}

	if p.Region == nil {
		errs = multierror.Append(errs, fmt.Errorf("snowflake provider region must be set in %s", component))
	}

	return errs
}

func (o *OktaProvider) Validate(component string) error {
	var errs *multierror.Error
	if o == nil {
		return nil
	}
	if o.OrgName == nil {
		errs = multierror.Append(errs, fmt.Errorf("okta provider org_name required in %s", component))
	}
	return errs
}

func (c *Config) ValidateSnowflakeProviders() error {
	var errs *multierror.Error
	c.WalkComponents(func(component string, comms ...Common) {
		p := ResolveSnowflakeProvider(comms...)
		if e := p.Validate(component); e != nil {
			errs = multierror.Append(errs, e)
		}
	})
	return errs.ErrorOrNil()
}

func (c *Config) ValidateOktaProviders() error {
	var errs *multierror.Error
	c.WalkComponents(func(component string, comms ...Common) {
		p := ResolveOktaProvider(comms...)
		if err := p.Validate(component); err != nil {
			errs = multierror.Append(errs, err)
		}
	})
	return errs
}

func (c *Config) ValidateBlessProviders() error {
	var errs *multierror.Error
	c.WalkComponents(func(component string, comms ...Common) {
		p := ResolveBlessProvider(comms...)
		if err := p.Validate(component); err != nil {
			errs = multierror.Append(errs, err)
		}
	})
	return errs
}

func (c *Config) ValidateGenericProviders() error {
	var errs *multierror.Error
	c.WalkComponents(func(component string, comms ...Common) {
		for name, p := range ResolveRequiredProviders(comms...) {
			if err := p.Validate(name, component); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	})
	return errs.ErrorOrNil()
}

func (c *Config) ValidateTravis() error {
	var errs *multierror.Error
	c.WalkComponents(func(component string, comms ...Common) {
		t := ResolveTravis(comms...)
		if t.Enabled == nil || !*t.Enabled {
			return // nothing to do
		}

		if t.AWSIAMRoleName == nil || *t.AWSIAMRoleName == "" {
			errs = multierror.Append(errs, fmt.Errorf("if travis is enabled, aws_role_name must be set"))
		}

		if t.Command != nil {
			_, ok := validCICommands[*t.Command]
			if !ok {
				errs = multierror.Append(errs, fmt.Errorf("unrecognized travisci command %s (%s)", *t.Command, component))
			}
		}
	})

	return errs
}

func (c *Config) validateTFE() error {
	var errs *multierror.Error

	if c.TFE != nil && c.TFE.TFEOrg == "" {
		errs = multierror.Append(errs, errors.Errorf("tfe org is required"))
	}

	return errs
}

func (c *Config) ValidateGithubActionsCI() error {
	var errs *multierror.Error
	c.WalkComponents(func(component string, comms ...Common) {
		t := ResolveGitHubActionsCI(comms...)
		if t.Enabled == nil || !*t.Enabled {
			return // nothing to do
		}

		if t.AWSIAMRoleName != nil && *t.AWSIAMRoleName != "" {
			if t.AWSRegion == nil || *t.AWSRegion == "" {
				errs = multierror.Append(errs, fmt.Errorf("if github_actions_ci.aws_role_name is set, aws_region must be set"))
			}
		}

		if t.Command != nil {
			_, ok := validCICommands[*t.Command]
			if !ok {
				errs = multierror.Append(errs, fmt.Errorf("unrecognized github_actions_ci command %s (%s)", *t.Command, component))
			}
		}
	})

	return errs
}

func (c *Config) ValidateToolsTfLint() ([]string, error) {
	var warns []string
	var errs *multierror.Error
	c.WalkComponents(func(component string, comms ...Common) {
		c := comms[len(comms)-1]
		if c.Tools != nil && c.Tools.TfLint != nil {
			warns = append(warns, fmt.Sprintf("per-component tflint config is not implemented, ignoring config in %s", component))
		}
	})

	return warns, errs
}
func (c *Config) WalkComponents(f func(component string, commons ...Common)) {
	for name, acct := range c.Accounts {
		f(fmt.Sprintf("accounts/%s", name), c.Defaults.Common, acct.Common)
	}

	f("global", c.Defaults.Common, c.Global.Common)

	for envName, env := range c.Envs {
		for componentName, component := range env.Components {
			f(fmt.Sprintf("%s/%s", envName, componentName), c.Defaults.Common, env.Common, component.Common)
		}
	}
}

// validateInheritedStringField will walk all accounts and components and ensure that a given field is valid at at least
// one level of the inheritance hierarchy. We should eventually distinuish between not present and invalid because
// if the value is present but invalid we should probably mark it as such, rather than papering over it.
func (c *Config) validateInheritedStringField(fieldName string, getter func(Common) *string, validator func(*string) bool) *multierror.Error {
	var err *multierror.Error

	// For each account, we need the field to be valid in either the defaults or account
	for acctName, acct := range c.Accounts {
		v := lastNonNil(getter, c.Defaults.Common, acct.Common)
		if !validator(v) {
			err = multierror.Append(err, fmt.Errorf("account %s must have a valid %s set at either the account or defaults level", acctName, fieldName))
		}
	}

	// global
	v := lastNonNil(getter, c.Defaults.Common, c.Global.Common)
	if !validator(v) {
		err = multierror.Append(err, fmt.Errorf("global must have a valid %s set at either the global or defaults level", fieldName))
	}

	// For each component, we need the field to be valid at one of defaults, env or component
	for envName, env := range c.Envs {
		for componentName, component := range env.Components {
			v := lastNonNil(getter, c.Defaults.Common, env.Common, component.Common)
			if !validator(v) {
				err = multierror.Append(err, fmt.Errorf("componnent %s/%s must have a valid %s", envName, componentName, fieldName))
			}
		}
	}

	return err
}

// validateExtraVars make sure users don't specify reserved variables
func (c *Config) validateExtraVars() error {
	var err *multierror.Error
	validate := func(extraVars map[string]string) {
		for extraVar := range extraVars {
			if _, ok := ReservedVariableNames[extraVar]; ok {
				err = multierror.Append(err, fmt.Errorf("extra_var[%s] is a fogg reserved variable name", extraVar))
			}
		}
	}
	extraVars := []map[string]string{}
	if c.Defaults.ExtraVars != nil {
		extraVars = append(extraVars, c.Defaults.ExtraVars)
	}
	for _, env := range c.Envs {
		extraVars = append(extraVars, env.ExtraVars)
		for _, component := range env.Components {
			extraVars = append(extraVars, component.ExtraVars)
		}
	}

	for _, acct := range c.Accounts {
		extraVars = append(extraVars, acct.ExtraVars)
	}

	for _, extraVar := range extraVars {
		validate(extraVar)
	}

	if err.ErrorOrNil() != nil {
		return errs.WrapUser(err.ErrorOrNil(), "extra_vars contains reserved variable names")
	}
	return nil
}

func (c *Config) validateAWSProviderAuth() error {
	var err *multierror.Error
	validate := func(p *AWSProvider) {
		if p != nil && (p.Profile != nil && p.Role != nil) {
			err = multierror.Append(err, fmt.Errorf("aws provider must have only one of profile or role set"))
		}
	}

	if c.Defaults.Providers != nil {
		validate(c.Defaults.Providers.AWS)
	}

	if c.Global.Providers != nil {
		validate(c.Global.Providers.AWS)
	}

	for _, env := range c.Envs {
		if env.Providers != nil {
			validate(env.Providers.AWS)
		}
		for _, component := range env.Components {
			if component.Providers != nil {
				validate(component.Providers.AWS)
			}
		}
	}

	for _, acct := range c.Accounts {
		if acct.Providers != nil {
			validate(acct.Providers.AWS)
		}
	}

	return err.ErrorOrNil()
}

func (c *Config) validateModules() error {
	minTFVersion := goVersion.Must(goVersion.NewVersion("0.12.0"))

	for name, module := range c.Modules {
		version := ResolveModuleTerraformVersion(c.Defaults, module)
		if version == nil {
			return fmt.Errorf("must set terrform version for module %s at either defaults or module level", name)
		}

		v, err := goVersion.NewVersion(*version)
		if err != nil {
			return errs.WrapUserf(err, "Could not parse semver terraform version [%s]", *version)
		}

		if v.LessThan(minTFVersion) {
			return errs.NewUserf("fogg only supports tf versions >= %s, but %s was provided", minTFVersion.String(), *version)
		}
	}
	return nil
}

func nonEmptyString(s *string) bool {
	return s != nil && len(*s) > 0
}

func (c *Config) ValidateFileDependencies(fs afero.Fs) error {
	var errs *multierror.Error
	c.WalkComponents(func(component string, comms ...Common) {
		files := ResolveOptionalStringSlice(DependsOnFilesGetter, comms...)
		keys := make(map[string]bool)
		for _, file := range files {
			ext := filepath.Ext(file)
			filename := filepath.Base(file)
			key := strings.TrimSuffix(filename, ext)
			key = util.ConvertToSnake(key)
			if keys[key] {
				errs = multierror.Append(errs, fmt.Errorf("component: %s - local.%s, file dependency naming collision. %v\n", component, key, files))
			} else {
				keys[key] = true
			}
			if _, err := fs.Stat(file); os.IsNotExist(err) {
				errs = multierror.Append(errs, fmt.Errorf("component: %s - File does not exist: %s\n", component, file))
			}
		}
	})

	return errs.ErrorOrNil()
}
