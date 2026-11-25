import type {
  AwsProviderConfig,
  Backend,
  CloudflareProviderConfig,
  ComponentConfig,
  ComponentConfigSource,
} from '@tcons/component'
import type {
  AWSProvider,
  Component,
  Backend as FoggBackend,
} from './imports/fogg-types.generated'
import { loadComponentConfig } from './util/load-component-config'

/** The default fogg component configuration file (relative to the cwd). */
export const FOGG_COMPONENT_FILE = '.fogg-component.yaml'

export interface FoggComponentMappingOptions {
  /**
   * Declare every `component_backends` entry as a dependency, even when fogg
   * did not filter them (`component_backends_filtered: false`).
   *
   * @default false - dependencies are only declared when `component_backends_filtered` is true
   */
  readonly forceRemoteStates?: boolean
}

export interface FoggYamlSourceOptions extends FoggComponentMappingOptions {
  /**
   * Path of the fogg component configuration.
   *
   * @default '.fogg-component.yaml' (in the current working directory)
   */
  readonly path?: string
}

/**
 * Reads a fogg generated `.fogg-component.yaml` and maps it to a
 * tool-neutral `ComponentConfig`. The fogg document is kept as `raw`
 * (exposed as `foggComp` on the fogg stacks).
 */
export class FoggYamlSource implements ComponentConfigSource {
  constructor(private readonly options: FoggYamlSourceOptions = {}) {}

  public load(): ComponentConfig {
    return foggComponentToConfig(
      loadComponentConfig(this.options.path ?? FOGG_COMPONENT_FILE),
      this.options,
    )
  }
}

function mapBackend(b: FoggBackend): Backend {
  if (b?.kind === 's3' && b.s3) {
    return {
      type: 's3',
      bucket: b.s3.bucket,
      key: b.s3.key_path,
      region: b.s3.region,
      dynamodbTable: b.s3.dynamo_table,
      profile: b.s3.profile,
      roleArn: b.s3.role_arn,
    }
  }
  // unsupported backend kinds are rejected by the wiring
  // ("Unsupported backend configuration <kind>"), as in v5
  return { type: b?.kind } as unknown as Backend
}

function mapAwsProvider(p: AWSProvider): AwsProviderConfig {
  return {
    region: p.region,
    alias: p.alias,
    accountId: p.account_id,
    profile: p.profile,
    roleArn: p.role_arn,
    defaultTags: p.default_tags?.enabled ? true : undefined,
    ignoreTags: p.ignore_tags?.enabled
      ? { keys: p.ignore_tags.keys, keyPrefixes: p.ignore_tags.key_prefixes }
      : undefined,
  }
}

/**
 * Map a fogg `Component` (`.fogg-component.yaml`) to a `ComponentConfig`.
 *
 * - dependencies are only declared when `forceRemoteStates` or
 *   `component_backends_filtered` is set
 * - only the `cloudflare/cloudflare` generic provider is supported, any
 *   other enabled `required_providers` entry throws
 * - the default AWS provider's `default_tags.tags` become `tags`, so they
 *   apply to every AWS provider (including regional ones)
 *
 * @throws on unsupported generic providers
 */
export function foggComponentToConfig(
  component: Component,
  options: FoggComponentMappingOptions = {},
): ComponentConfig {
  const pc =
    component.providers_configuration ??
    ({} as Component['providers_configuration'])

  let cloudflare: CloudflareProviderConfig | undefined
  for (const p of Object.values(component.required_providers ?? {})) {
    if (!p.enabled) continue
    switch (p.source) {
      case 'cloudflare/cloudflare':
        cloudflare = { config: p.config }
        break
      default:
        throw new Error(`Unsupported provider ${p.source}`)
    }
  }

  const withDependencies =
    (options.forceRemoteStates ?? false) ||
    component.component_backends_filtered

  // v5 defaulted (and wrote back) module_name when a module_source is set
  const raw: Component = component.module_source
    ? { ...component, module_name: component.module_name ?? 'main' }
    : component

  return {
    identity: {
      name: component.name,
      env: component.env,
      project: component.project,
      owner: component.owner,
    },
    backend: mapBackend(component.backend),
    tags: pc.aws?.default_tags?.enabled ? pc.aws.default_tags.tags : undefined,
    providers: {
      aws: pc.aws ? mapAwsProvider(pc.aws) : undefined,
      awsRegional: (pc.aws_regional_providers ?? []).map(mapAwsProvider),
      datadog: pc.datadog ? {} : undefined,
      cloudflare,
    },
    dependencies: withDependencies
      ? Object.fromEntries(
          Object.entries(component.component_backends ?? {}).map(
            ([name, backend]) => [name, mapBackend(backend)],
          ),
        )
      : {},
    locals: component.locals_block,
    mainModule: component.module_source
      ? { source: component.module_source, name: raw.module_name }
      : undefined,
    modules: component.modules,
    raw,
  }
}

/**
 * The fogg `Component` a config was mapped from (`config.raw`).
 *
 * @throws if the config was not produced by {@link foggComponentToConfig}
 */
export function foggComponentOf(config: ComponentConfig): Component {
  if (!config.raw) {
    throw new Error(
      'foggComp is not available: the component config has no `raw` fogg document (it was not loaded through FoggYamlSource / foggComponentToConfig)',
    )
  }
  return config.raw as Component
}
