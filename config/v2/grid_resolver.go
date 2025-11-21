package v2

func ResolveGrid(commons ...Common) *GridConfig {
	var enabled *bool
	var guid *string
	var endpoint *string
	var project *string
	var envMap map[string]string
	var accountMap map[string]string

	for _, c := range commons {
		if c.Grid != nil {
			if c.Grid.Enabled != nil {
				enabled = c.Grid.Enabled
			}
			if c.Grid.GUID != nil {
				guid = c.Grid.GUID
			}
			if c.Grid.Endpoint != nil {
				endpoint = c.Grid.Endpoint
			}
			if c.Grid.Project != nil {
				project = c.Grid.Project
			}
			if c.Grid.EnvMap != nil {
				envMap = c.Grid.EnvMap
			}
			if c.Grid.AccountMap != nil {
				accountMap = c.Grid.AccountMap
			}
		}
	}

	if enabled == nil && guid == nil && endpoint == nil && project == nil && envMap == nil && accountMap == nil {
		return nil
	}

	return &GridConfig{
		Enabled:    enabled,
		GUID:       guid,
		Endpoint:   endpoint,
		Project:    project,
		EnvMap:     envMap,
		AccountMap: accountMap,
	}
}

// ResolveLogicalIDProject returns the project name to use for logical IDs
// Uses grid.project override if set, otherwise uses actual project name
func ResolveLogicalIDProject(gridConfig *GridConfig, actualProject string) string {
	if gridConfig != nil && gridConfig.Project != nil && *gridConfig.Project != "" {
		return *gridConfig.Project
	}
	return actualProject
}

// ResolveLogicalIDEnv returns the env name to use for logical IDs
// Uses grid.env_map mapping if set, otherwise uses actual env name
func ResolveLogicalIDEnv(gridConfig *GridConfig, actualEnv string) string {
	if gridConfig != nil && gridConfig.EnvMap != nil {
		if shortName, ok := gridConfig.EnvMap[actualEnv]; ok && shortName != "" {
			return shortName
		}
	}
	return actualEnv
}

// ResolveLogicalIDAccount returns the account name to use for logical IDs
// Uses grid.account_map mapping if set, otherwise uses actual account name
func ResolveLogicalIDAccount(gridConfig *GridConfig, actualAccount string) string {
	if gridConfig != nil && gridConfig.AccountMap != nil {
		if shortName, ok := gridConfig.AccountMap[actualAccount]; ok && shortName != "" {
			return shortName
		}
	}
	return actualAccount
}
