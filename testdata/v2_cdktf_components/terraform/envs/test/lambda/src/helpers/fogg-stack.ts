// Auto-generated by fogg. Do not edit
// Make improvements in fogg, so that everyone can benefit.

import * as fs from "fs";
import { provider as awsProvider } from "@cdktf/provider-aws";
import { provider as cloudflareProvider } from "@cdktf/provider-cloudflare";
import { provider as dataDogProvider } from "@cdktf/provider-datadog";
import {
  TerraformStack,
  S3Backend,
  S3BackendConfig,
  DataTerraformRemoteStateS3,
  DataTerraformRemoteStateS3Config,
  TerraformHclModule,
  TerraformLocal,
} from "cdktf";
import { Construct } from "constructs";
import * as yaml from "js-yaml";

import {
  Component,
  Backend,
  AWSProvider,
  DatadogProvider,
  GenericProvider,
} from "./fogg-types.generated";

export interface FoggStackProps {
  /**
   * Force remote state configuration
   * @default false - only configure remote states if `component_backends_filtered` is true
   */
  forceRemoteStates?: boolean;
}

export function loadComponentConfig(): Component {
  const file = fs.readFileSync(`.fogg-component.yaml`, "utf8");
  const componentConfig = yaml.load(file) as Component;
  return replaceNullWithUndefined(componentConfig);
}

/**
 * Helper stack to wrap Fogg component configuration and set up configured providers and backends.
 */
export class FoggStack extends TerraformStack {
  public readonly foggComp: Component;
  public readonly modules: Record<string, TerraformHclModule> = {};
  public readonly locals: Record<string, TerraformLocal> = {};

  constructor(scope: Construct, id: string, props: FoggStackProps) {
    super(scope, id);
    this.foggComp = loadComponentConfig();

    this.parseBackendConfig();
    this.parseBundledProviderConfig();
    for (const p of Object.values(this.foggComp.required_providers)) {
      if (p.enabled) this.parseGenericProviderConfig(p);
    }
    // parse remote backends
    const forceRemoteBackend = props.forceRemoteStates ?? false;
    if (forceRemoteBackend || this.foggComp.component_backends_filtered) {
      for (const [name, remoteStateConfig] of Object.entries(
        this.foggComp.component_backends,
      )) {
        this.parseRemoteState(name, remoteStateConfig);
      }
    }
    this.parseLocalsBlock();
    this.parseModules();
  }

  /**
   * Set variables for the main module defined in the fogg component configuration.
   *
   * @param variables - The variables to set for the module
   */
  public setMainModuleVariables(variables: Record<string, any>): void {
    const id = (this.foggComp.module_name =
      this.foggComp.module_name ?? "main");
    this.setModuleVariables(id, variables);
  }

  /**
   * Set variables for a module included in the fogg component modules[] configuration.
   *
   * @param name - The module name as defined in the fogg component configuration
   * @param variables - The variables to set for the module
   */
  public setModuleVariables(
    name: string,
    variables: Record<string, any>,
  ): void {
    if (!this.modules[name]) {
      throw new Error(`Module ${name} not found`);
    }
    for (const [key, value] of Object.entries(variables)) {
      this.modules[name].set(key, value);
    }
  }

  /**
   * Get a local defined in the fogg component configuration.
   *
   * @param name the name of the local to get
   * @returns the TerraformLocal object
   */
  public getLocal(name: string): TerraformLocal {
    if (!this.locals[name]) {
      throw new Error(`Local ${name} not found`);
    }
    return this.locals[name];
  }

  private parseBackendConfig(): void {
    if (this.foggComp.backend.kind === "s3" && this.foggComp.backend.s3) {
      const s3Config = this.foggComp.backend.s3;
      let s3BackendConfig: Mutable<S3BackendConfig> = {
        bucket: s3Config.bucket,
        dynamodbTable: s3Config.dynamo_table,
        key: s3Config.key_path,
        region: s3Config.region,
        encrypt: true,
      };
      if (s3Config.profile) {
        s3BackendConfig.profile = s3Config.profile;
      } else if (s3Config.role_arn) {
        s3BackendConfig.assumeRole = {
          roleArn: s3Config.role_arn,
        };
      }
      // console.log(
      //   `Setting S3 backend Config ${JSON.stringify(s3BackendConfig, null, 2)}`
      // );
      new S3Backend(this, s3BackendConfig);
    } else {
      throw new Error(
        `Unsupported backend configuration ${this.foggComp.backend.kind}`,
      );
    }
  }

  private parseRemoteState(id: string, remoteConfig: Backend): void {
    if (remoteConfig.kind === "s3" && remoteConfig.s3) {
      const s3Config = remoteConfig.s3;
      let remoteStateConfig: Mutable<DataTerraformRemoteStateS3Config> = {
        bucket: s3Config.bucket,
        dynamodbTable: s3Config.dynamo_table,
        key: s3Config.key_path,
        region: s3Config.region,
        encrypt: true,
      };
      if (s3Config.profile) {
        remoteStateConfig.profile = s3Config.profile;
      } else if (s3Config.role_arn) {
        remoteStateConfig.assumeRole = {
          roleArn: s3Config.role_arn,
        };
      }
      // console.log(
      //   `Setting ${id} Remote backend Config ${JSON.stringify(
      //     remoteStateConfig,
      //     null,
      //     2
      //   )}`
      // );
      new DataTerraformRemoteStateS3(this, id, remoteStateConfig);
    } else {
      throw new Error(`Unsupported backend configuration ${remoteConfig.kind}`);
    }
  }

  private parseBundledProviderConfig(): void {
    const providers = this.foggComp.providers_configuration;
    if (providers.aws) {
      this.parseAwsProviderConfig(providers.aws);
    }
    if (providers.aws_regional_providers) {
      for (let i = 0; i < providers.aws_regional_providers.length; i++) {
        this.parseAwsProviderConfig(
          providers.aws_regional_providers[i],
          `aws-${i}`,
        );
      }
    }
    if (providers.datadog) {
      this.parseDataDogProviderConfig(providers.datadog);
    }
  }

  private parseGenericProviderConfig(config: GenericProvider): void {
    switch (config.source) {
      case "cloudflare/cloudflare":
        this.parseCloudflareProviderConfig(config);
        break;
      default:
        throw new Error(`Unsupported provider ${config.source}`);
    }
  }

  private parseLocalsBlock() {
    if (this.foggComp.locals_block) {
      for (const [key, value] of Object.entries(this.foggComp.locals_block)) {
        this.locals[key] = new TerraformLocal(this, key, `\${${value}}`);
      }
    }
  }

  private parseModules() {
    if (this.foggComp.module_source) {
      const id = (this.foggComp.module_name =
        this.foggComp.module_name ?? "main");
      this.modules[id] = new TerraformHclModule(this, id, {
        source: this.foggComp.module_source,
      });
    }

    for (let i = 0; i < this.foggComp.modules.length; i++) {
      const moduleConfig = this.foggComp.modules[i];
      const id = moduleConfig.name ?? `module_${i}`;
      if (!moduleConfig.source) {
        console.warn(`Module ${id} does not have a source, skipping`);
        continue;
      }
      if (this.modules[id]) {
        throw new Error(`Module ${id} already exists`);
      }
      if (!moduleConfig.name) {
        console.log(
          `Module ${moduleConfig.source} does not have a name, using ${id}`,
        );
      }
      this.modules[id] = new TerraformHclModule(this, id, {
        source: moduleConfig.source,
        version: moduleConfig.version,
      });
      // TODO: Add validation for module variables
      // TODO: Export module outputs
    }
  }

  private parseAwsProviderConfig(
    config: AWSProvider,
    id: string = "Default",
  ): void {
    const c: Mutable<awsProvider.AwsProviderConfig> = {
      region: config.region,
      alias: config.alias,
    };
    if (config.profile) {
      c.profile = config.profile;
    } else if (config.role_arn) {
      c.assumeRole = [
        {
          roleArn: config.role_arn,
        },
      ];
    }
    new awsProvider.AwsProvider(this, id, c);
  }

  private parseDataDogProviderConfig(_config: DatadogProvider): void {
    new dataDogProvider.DatadogProvider(this, "datadog", {});
  }

  private parseCloudflareProviderConfig(_config: GenericProvider): void {
    new cloudflareProvider.CloudflareProvider(this, "cloudflare", {});
  }
}

// helper type to make readonly interface properties mutable
type Mutable<T> = {
  -readonly [P in keyof T]: T[P];
};

// helper function to replace fogg "null" values with "undefined"
function replaceNullWithUndefined(obj: any): any {
  if (obj === null) {
    return undefined;
  }
  if (Array.isArray(obj)) {
    return obj.map(replaceNullWithUndefined);
  }
  if (typeof obj === "object" && obj !== null) {
    const newObj: any = {};
    for (const key in obj) {
      newObj[key] = replaceNullWithUndefined(obj[key]);
    }
    return newObj;
  }
  return obj;
}
