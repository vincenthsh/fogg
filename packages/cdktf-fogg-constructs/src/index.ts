export * from './fogg-stack'
export * from './fogg-terrastack'
export * from './fogg-yaml-source'
export type { FoggRemoteStateSchema } from './util/remote-state'

/**
 * Re-exports from `@tcons/component` (the shared component core), so
 * consumers only depend on this package.
 *
 * `Backend` / `S3Backend` are re-exported as `ComponentBackend` /
 * `ComponentS3Backend` to avoid clashing with the generated fogg types.
 */
export {
  type AwsProviderConfig,
  type Backend as ComponentBackend,
  type CloudflareProviderConfig,
  type ComponentConfig,
  type ComponentConfigInput,
  type ComponentConfigSource,
  type ComponentIdentity,
  ComponentStack,
  type ComponentStackProps,
  type ComponentTagOptions,
  ComponentWiring,
  type ComponentWiringOptions,
  DEFAULT_AWS_PROVIDER_ID,
  type IComponentWiring,
  type IRemoteStateAccess,
  type LegacyRemoteStateSchema,
  type ListOutput,
  type MainModuleConfig,
  type MapOutput,
  type ModuleConfig,
  type ProvidersConfig,
  type RemoteOutput,
  type RemoteStateOutputs,
  type RemoteStateSchema,
  type S3Backend as ComponentS3Backend,
  type StringifiedRecord,
  wireComponent,
} from '@tcons/component'
export {
  ComponentAwsStack,
  type ComponentAwsStackProps,
} from '@tcons/component/aws'
export {
  fromLegacySchema,
  type InferOutputSchema,
  type LegacyOutputType,
} from '@tcons/component/legacy'

// export testing helpers
export {
  setupSchemaMatchers,
  toMatchSchema,
  type SchemaVitestMatchers,
} from './matchers/schema'

/**
 * Auto generated types
 *
 * update with pnpm run go:generate
 */
export * from './imports/fogg-types.generated'
