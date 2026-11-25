import type { provider as awsProvider } from '@cdktn/provider-aws'
import type {
  LegacyRemoteStateSchema,
  RemoteStateOutputs,
  RemoteStateSchema,
  StringifiedRecord,
} from '@tcons/component'
import { ComponentAwsStack } from '@tcons/component/aws'
import type { InferOutputSchema } from '@tcons/component/legacy'
import type { Construct } from 'constructs'
import type { AwsStackProps } from 'terraconstructs/lib/aws'
import {
  type FoggStackProps,
  type IFoggStack,
  resolveFoggStackConfig,
} from './fogg-stack'
import { foggComponentOf } from './fogg-yaml-source'
import type { Component } from './imports/fogg-types.generated'
import {
  type FoggRemoteStateSchema,
  toRemoteStateSchema,
} from './util/remote-state'

export interface FoggTerraStackProps
  extends FoggStackProps,
    Omit<AwsStackProps, 'providerConfig'> {}

function toComponentAwsStackProps(props: FoggTerraStackProps) {
  const {
    forceRemoteStates: _forceRemoteStates,
    configPath: _configPath,
    source: _source,
    config: _config,
    ...awsStackProps
  } = props
  return { ...awsStackProps, config: resolveFoggStackConfig(props) }
}

/**
 * Helper AwsStack to wrap Fogg component configuration and set up configured providers and backends.
 *
 * A thin wrapper over `@tcons/component/aws`'s `ComponentAwsStack`,
 * configured from `.fogg-component.yaml` (or the `source` / `config` props).
 */
export class FoggTerraStack extends ComponentAwsStack implements IFoggStack {
  constructor(scope: Construct, id: string, props: FoggTerraStackProps) {
    super(scope, id, toComponentAwsStackProps(props))
  }

  /**
   * The Raw Fogg component configuration (`componentConfig.raw`).
   *
   * @throws if the config was not loaded from a fogg component document
   */
  public get foggComp(): Component {
    return foggComponentOf(this.componentConfig)
  }

  /**
   * Return an AWS provider by alias (`DefaultAwsProvider` for the default one).
   * @throws if the provider is not found
   */
  public awsProvider(alias: string): awsProvider.AwsProvider {
    return this.componentWiring.awsProvider(alias)
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
