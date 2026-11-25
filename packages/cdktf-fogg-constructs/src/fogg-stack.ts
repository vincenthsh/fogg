import type { provider as awsProvider } from '@cdktn/provider-aws'
import {
  type ComponentConfig,
  type ComponentConfigSource,
  ComponentStack,
  type LegacyRemoteStateSchema,
  type RemoteStateOutputs,
  type RemoteStateSchema,
  type StringifiedRecord,
} from '@tcons/component'
import type { InferOutputSchema } from '@tcons/component/legacy'
import type { TerraformHclModule, TerraformLocal } from 'cdktn'
import type { Construct, IConstruct } from 'constructs'
import { FoggYamlSource, foggComponentOf } from './fogg-yaml-source'
import type { Component } from './imports/fogg-types.generated'
import {
  type FoggRemoteStateSchema,
  toRemoteStateSchema,
} from './util/remote-state'

export interface FoggStackProps {
  /**
   * Force remote state configuration
   *
   * Only applies when the configuration is read from `.fogg-component.yaml`
   * (neither `source` nor `config` is passed).
   *
   * @default false - only configure remote states if `component_backends_filtered` is true
   */
  forceRemoteStates?: boolean

  /**
   * Default tags to apply to resources on top of the default tags set by the Fogg component configuration
   *
   * @default - No additional tags
   */
  defaultTags?: Record<string, string>
  /**
   * Tags to ignore on top of the ignoreTags set by the Fogg component configuration
   */
  ignoreTags?: awsProvider.AwsProviderIgnoreTags

  /**
   * Path of the fogg component configuration file.
   *
   * @default '.fogg-component.yaml' (in the current working directory)
   */
  configPath?: string

  /**
   * Load the component configuration from this source instead of
   * `.fogg-component.yaml`. Mutually exclusive with `config`.
   *
   * @default - a `FoggYamlSource` for `configPath`
   */
  source?: ComponentConfigSource

  /**
   * Use this component configuration instead of loading
   * `.fogg-component.yaml`. Mutually exclusive with `source`.
   *
   * `foggComp` is only available when `config.raw` is the fogg document.
   * The resolved configuration is exposed as `componentConfig`.
   *
   * @default - loaded from `source`
   */
  config?: ComponentConfig
}

/**
 * Resolve the component configuration for the fogg stack props:
 * `config`, else `source.load()`, else `.fogg-component.yaml`.
 */
export function resolveFoggStackConfig(props: FoggStackProps): ComponentConfig {
  if (props.config && props.source) {
    throw new Error('Pass either `config` or `source`, not both')
  }
  if (props.config) return props.config
  if (props.source) return props.source.load()
  return new FoggYamlSource({
    path: props.configPath,
    forceRemoteStates: props.forceRemoteStates,
  }).load()
}

export interface IFoggStack extends IConstruct {
  /**
   * The Raw Fogg component configuration
   */
  readonly foggComp: Component
  /**
   * A Map of module name to TerraformHclModule Constructs parsed from the fogg component configuration
   */
  readonly modules: Record<string, TerraformHclModule>
  /**
   * A Map of local name to TerraformLocal Constructs parsed from the fogg component configuration
   */
  readonly locals: Record<string, TerraformLocal>
  /**
   * The AWS region defined in the aws providers configuration (if provided)
   * or undefined if not set
   */
  readonly region?: string
  /**
   * The default AWS provider
   */
  readonly defaultAwsProvider: awsProvider.AwsProvider
  /**
   * Set variables for the main module defined in the fogg component configuration.
   *
   * @param variables - The variables to set for the module
   */
  setMainModuleVariables(variables: Record<string, any>): void
  /**
   * Returns an object to access properties within a remote state defined in the fogg component configuration.
   * All properties will be returned as strings.
   *
   * @param name - The name of the remote state to get
   * @returns a Remote State Access proxy to transparently access the remote state outputs as strings
   * @throws if the remote state is not found
   */
  remoteState<T = Record<string, string>>(name: string): StringifiedRecord<T>
  /**
   * Returns an object to access properties within a remote state, typed by a
   * schema converted with `fromLegacySchema<T>()`.
   */
  remoteState<T>(name: string, schema: LegacyRemoteStateSchema<T>): T
  /**
   * Returns an object to access properties within a remote state, typed and
   * rendered per the zod schema.
   *
   * @throws if the remote state is not found
   */
  remoteState<S extends RemoteStateSchema>(
    name: string,
    schema: S,
  ): RemoteStateOutputs<S>
  /**
   * Returns an object to access properties within a remote state defined in the fogg component configuration.
   * Property types will be inferred from the provided (v5 string) schema.
   *
   * @param name - The name of the remote state to get
   * @param schema - The schema to use to infer the property types
   * @returns a Remote State Access proxy to transparently access the remote state outputs with typed properties
   * @throws if the remote state is not found
   */
  remoteState<T>(name: string, schema: InferOutputSchema<T>): T
  /**
   * Set variables for a module included in the fogg component modules[] configuration.
   *
   * @param name - The module name as defined in the fogg component configuration
   * @param variables - The variables to set for the module
   */
  setModuleVariables(name: string, variables: Record<string, any>): void
  /**
   * Get a local defined in the fogg component configuration.
   *
   * @param name the name of the local to get
   * @returns the TerraformLocal object
   */
  getLocal(name: string): TerraformLocal
}

/**
 * Helper stack to wrap Fogg component configuration and set up configured providers and backends.
 *
 * A thin wrapper over `@tcons/component`'s `ComponentStack`, configured from
 * `.fogg-component.yaml` (or the `source` / `config` props).
 */
export class FoggStack extends ComponentStack implements IFoggStack {
  constructor(scope: Construct, id: string, props: FoggStackProps = {}) {
    super(scope, id, {
      config: resolveFoggStackConfig(props),
      defaultTags: props.defaultTags,
      ignoreTags: props.ignoreTags,
    })
  }

  /**
   * The Raw Fogg component configuration (`componentConfig.raw`).
   *
   * @throws if the config was not loaded from a fogg component document
   */
  public get foggComp(): Component {
    return foggComponentOf(this.componentConfig)
  }

  public remoteState<T = Record<string, string>>(
    name: string,
  ): StringifiedRecord<T>
  public remoteState<T>(name: string, schema: LegacyRemoteStateSchema<T>): T
  public remoteState<S extends RemoteStateSchema>(
    name: string,
    schema: S,
  ): RemoteStateOutputs<S>
  public remoteState<T>(name: string, schema: InferOutputSchema<T>): T
  public remoteState(name: string, schema?: FoggRemoteStateSchema): unknown {
    return super.remoteState(name, toRemoteStateSchema(schema) as any)
  }
}
