import * as fs from 'node:fs'
import * as path from 'node:path'
import { App, TerraformOutput, Testing } from 'cdktn'
import * as yaml from 'js-yaml'
import { beforeEach, describe, expect, it } from 'vitest'
import { z } from 'zod'
import {
  type Component,
  type ComponentConfig,
  FoggStack,
  FoggTerraStack,
  FoggYamlSource,
  type InferOutputSchema,
  type ListOutput,
  type MapOutput,
  foggComponentToConfig,
  fromLegacySchema,
} from '../src'
import {
  IMPLS,
  givenComponent,
  setupParity,
  synthOrError,
} from './helpers/parity'

setupParity()

const FIXTURE = path.join(__dirname, 'util', 'fixtures', '.fogg-component.yaml')
const terraProps = { gridUUID: 'g123-456', environmentName: 'development' }

/** the full fixture, as written by fogg (nulls included) */
function useFixtureFile(): void {
  fs.copyFileSync(FIXTURE, '.fogg-component.yaml')
}

/** the full fixture, patched (re-dumped, so nulls stay nulls) */
function useFixture(patch: (c: any) => void): void {
  const c = yaml.load(fs.readFileSync(FIXTURE, 'utf8')) as any
  patch(c)
  fs.writeFileSync('.fogg-component.yaml', yaml.dump(c))
}

interface InternalZone {
  name: string
  id: string
}
interface NetworkOutputs {
  vpc_id: string
  az_count: number
  app_subnet_ids: ListOutput<string>
  internal_dns_zone: MapOutput<InternalZone>
}
const networkOutputsSchema: InferOutputSchema<NetworkOutputs> = {
  vpc_id: 'string',
  az_count: 'number',
  app_subnet_ids: 'list',
  internal_dns_zone: 'map',
}

function useOutputs(stack: FoggStack | FoggTerraStack, name: string): void {
  const out = stack.remoteState<NetworkOutputs>(name, networkOutputsSchema)
  const o = (id: string, value: any) =>
    new TerraformOutput(stack, `${name}_${id}`, { value, staticId: true })
  o('vpc_id', out.vpc_id)
  o('az_count', out.az_count)
  o('subnet', out.app_subnet_ids.element(0))
  o('zone_id', out.internal_dns_zone.lookup('id'))
  o('zone_name', out.internal_dns_zone.lookup('name', ''))
  o('nested', out.internal_dns_zone.lookupNested('a', 'b'))
  o('plain', stack.remoteState(name).anything)
}

describe.each(IMPLS)(
  'full .fogg-component.yaml fixture [$name]',
  ({ FoggStack, FoggTerraStack }) => {
    for (const forceRemoteStates of [false, true]) {
      it(`FoggStack (forceRemoteStates=${forceRemoteStates})`, () => {
        useFixtureFile()
        const stack = new FoggStack(Testing.app(), 'MyStack', {
          forceRemoteStates,
          defaultTags: { purpose: 'parity', owner: 'override' },
          ignoreTags: { keyPrefixes: ['kubernetes.io/'] },
        })
        stack.setModuleVariables('vpc', { name: 'vpc', cidr: '10.0.0.0/16' })
        if (forceRemoteStates) {
          for (const dep of ['app', 'lambda', 'network']) useOutputs(stack, dep)
        }
      })
    }

    it('FoggStack with component tags, profile backend and main module', () => {
      useFixture((c) => {
        c.backend.s3.profile = 'state-profile'
        c.component_backends.app.s3.profile = 'state-profile'
        c.providers_configuration.aws.default_tags.tags = {
          env: 'from-fogg',
          team: 'platform',
        }
        c.providers_configuration.aws.ignore_tags = {
          enabled: true,
          keys: ['a', 'b'],
          key_prefixes: ['x/'],
        }
        c.providers_configuration.aws_regional_providers[0].profile = 'p'
        c.providers_configuration.datadog = { enabled: true, version: '3.0' }
        c.required_providers = {
          cloudflare: {
            enabled: true,
            source: 'cloudflare/cloudflare',
            version: '~> 4.0',
            config: { api_token: 'tok' },
          },
          disabled: { enabled: false, source: 'foo/bar', config: {} },
        }
        c.module_source = 'terraform/modules/network'
        c.component_backends_filtered = true
      })
      const stack = new FoggStack(Testing.app(), 'MyStack', {
        defaultTags: { team: 'override' },
        ignoreTags: { keys: ['b', 'c'] },
      })
      stack.setMainModuleVariables({ foo: 'bar' })
      useOutputs(stack, 'lambda')
      expect(stack.foggComp.module_name).toBe('main')
    })

    for (const forceRemoteStates of [false, true]) {
      it(`FoggTerraStack (forceRemoteStates=${forceRemoteStates})`, () => {
        useFixture((c) => {
          c.providers_configuration.aws_regional_providers = []
          c.providers_configuration.aws.default_tags.tags = { team: 'p' }
        })
        const stack = new FoggTerraStack(Testing.app(), 'MyStack', {
          ...terraProps,
          forceRemoteStates,
          defaultTags: { purpose: 'parity' },
          ignoreTags: { keyPrefixes: ['kubernetes.io/'] },
        })
        stack.setModuleVariables('vpc', { name: 'vpc' })
        if (forceRemoteStates) useOutputs(stack, 'network')
      })
    }

    it('FoggTerraStack rejects regional providers', () => {
      useFixtureFile()
      expect(
        () => new FoggTerraStack(Testing.app(), 'MyStack', terraProps),
      ).toThrow(/AWS regional providers are not supported/)
    })
  },
)

describe('v6', () => {
  let app: App
  beforeEach(() => {
    app = Testing.app()
  })

  describe('remoteState with zod schemas', () => {
    const zodSchema = z.object({
      vpc_id: z.string(),
      az_count: z.number(),
      app_subnet_ids: z.array(z.string()),
      internal_dns_zone: z.object({ id: z.string(), name: z.string() }),
    })

    function build(
      Stack: typeof FoggStack,
      access: (s: FoggStack) => Record<string, unknown>,
    ): string {
      useFixtureFile()
      const stack = new Stack(Testing.app(), 'MyStack', {
        forceRemoteStates: true,
      })
      for (const [id, value] of Object.entries(access(stack))) {
        new TerraformOutput(stack, id, { value, staticId: true })
      }
      return synthOrError(stack)
    }

    const legacyAccess = (s: FoggStack) => {
      const o = s.remoteState<NetworkOutputs>('network', networkOutputsSchema)
      return {
        vpc_id: o.vpc_id,
        az_count: o.az_count,
        subnet: o.app_subnet_ids.element(0),
        zone_id: o.internal_dns_zone.lookup('id'),
        zone_name: o.internal_dns_zone.lookup('name', ''),
        undeclared: (o as any).other,
      }
    }

    it('renders identically to the v5.0.1 legacy string schema', () => {
      const viaZod = build(FoggStack, (s) => {
        const o = s.remoteState('network', zodSchema)
        // typed accessors
        const subnet: string = o.app_subnet_ids.element(0)
        const zoneId: string = o.internal_dns_zone.lookup('id')
        return {
          vpc_id: o.vpc_id,
          az_count: o.az_count,
          subnet,
          zone_id: zoneId,
          zone_name: o.internal_dns_zone.lookup('name', ''),
          undeclared: (o as any).other,
        }
      })
      const viaLegacyV6 = build(FoggStack, legacyAccess)
      const viaLegacyV5 = build(IMPLS[0].FoggStack, legacyAccess)
      expect(viaZod).not.toMatch(/^ERROR/)
      expect(viaZod).toBe(viaLegacyV5)
      expect(viaLegacyV6).toBe(viaLegacyV5)
    })

    it('accepts fromLegacySchema() schemas', () => {
      const viaFromLegacy = build(FoggStack, (s) => {
        const o = s.remoteState(
          'network',
          fromLegacySchema<NetworkOutputs>(networkOutputsSchema),
        )
        return {
          vpc_id: o.vpc_id,
          az_count: o.az_count,
          subnet: o.app_subnet_ids.element(0),
          zone_id: o.internal_dns_zone.lookup('id'),
          zone_name: o.internal_dns_zone.lookup('name', ''),
          undeclared: (o as any).other,
        }
      })
      expect(viaFromLegacy).toBe(build(FoggStack, legacyAccess))
    })

    it('works on FoggTerraStack', () => {
      useFixture((c) => {
        c.providers_configuration.aws_regional_providers = []
      })
      const stack = new FoggTerraStack(app, 'MyStack', {
        ...terraProps,
        forceRemoteStates: true,
      })
      const o = stack.remoteState('network', zodSchema)
      expect(stack.resolve(o.internal_dns_zone.lookup('id'))).toBe(
        '${data.terraform_remote_state.network.outputs.internal_dns_zone.id}',
      )
    })

    it('rejects unknown legacy output types', () => {
      useFixtureFile()
      const stack = new FoggStack(app, 'MyStack', { forceRemoteStates: true })
      expect(() =>
        stack.remoteState('network', { vpc_id: 'set' } as any),
      ).toThrow(/Unsupported legacy output type "set"/)
    })

    it('throws for undeclared dependencies', () => {
      useFixtureFile()
      const stack = new FoggStack(app, 'MyStack')
      expect(() => stack.remoteState('network', zodSchema)).toThrow(
        /Remote state network not found/,
      )
    })
  })

  describe('configuration inputs', () => {
    it('reads configPath', () => {
      fs.mkdirSync('nested')
      fs.copyFileSync(FIXTURE, 'nested/component.yaml')
      const stack = new FoggStack(app, 'MyStack', {
        configPath: 'nested/component.yaml',
      })
      expect(stack.foggComp.name).toBe('network')
      expect(stack.componentConfig.identity).toEqual({
        name: 'network',
        env: 'test',
        project: 'proj',
        owner: 'foo@example.com',
      })
    })

    it('accepts a source', () => {
      const source = new FoggYamlSource({
        path: FIXTURE,
        forceRemoteStates: true,
      })
      const stack = new FoggStack(app, 'MyStack', { source })
      expect(stack.remoteState('lambda')).toBeDefined()
      expect(stack.foggComp.component_backends_filtered).toBe(false)
    })

    it('accepts a config, synthesizing like the yaml file', () => {
      useFixtureFile()
      const fromFile = new FoggStack(Testing.app(), 'MyStack')
      const config = new FoggYamlSource({ path: FIXTURE }).load()
      const fromConfig = new FoggStack(Testing.app(), 'MyStack', { config })
      expect(synthOrError(fromConfig)).toBe(synthOrError(fromFile))
    })

    it('accepts a tool-neutral config without fogg document', () => {
      const config: ComponentConfig = {
        identity: { name: 'n', env: 'e', project: 'p', owner: 'o' },
        backend: { type: 's3', bucket: 'b', key: 'k', region: 'r' },
        providers: { aws: { region: 'r', defaultTags: true } },
      }
      const stack = new FoggStack(app, 'MyStack', { config })
      expect(stack.defaultAwsProvider).toBeDefined()
      expect(() => stack.foggComp).toThrow(/foggComp is not available/)
    })

    it('lets subclasses declare their own config / component members', () => {
      useFixtureFile()
      class Mine extends FoggStack {
        public readonly config = { mine: true }
        public component = 'mine'
        constructor(scope: App, id: string) {
          super(scope, id, { forceRemoteStates: true })
        }
      }
      const reference = new FoggStack(Testing.app(), 'MyStack', {
        forceRemoteStates: true,
      })
      const stack = new Mine(Testing.app(), 'MyStack')
      expect(stack.config).toEqual({ mine: true })
      expect(stack.component).toBe('mine')
      expect(stack.foggComp.name).toBe('network')
      expect(stack.componentConfig.identity.name).toBe('network')
      expect(stack.remoteState('lambda')).toBeDefined()
      expect(synthOrError(stack)).toBe(synthOrError(reference))
    })

    it('rejects config together with source', () => {
      const config = new FoggYamlSource({ path: FIXTURE }).load()
      expect(
        () =>
          new FoggStack(app, 'MyStack', {
            config,
            source: { load: () => config },
          }),
      ).toThrow(/either `config` or `source`/)
    })

    it('passes AwsStack props through FoggTerraStack', () => {
      useFixture((c) => {
        c.providers_configuration.aws_regional_providers = []
      })
      const stack = new FoggTerraStack(app, 'MyStack', terraProps)
      expect(stack.gridUUID).toBe(terraProps.gridUUID)
      expect(stack.awsProvider('DefaultAwsProvider')).toBe(
        stack.defaultAwsProvider,
      )
      expect(() => stack.awsProvider('nope')).toThrow(/Provider nope not found/)
    })
  })

  describe('foggComponentToConfig', () => {
    const base = (): Component =>
      new FoggYamlSource({ path: FIXTURE }).load().raw as Component

    it('maps the fogg document', () => {
      const config = foggComponentToConfig(base(), { forceRemoteStates: true })
      expect(config.backend).toMatchObject({
        type: 's3',
        bucket: 'buck',
        key: 'terraform/proj/envs/test/components/network.tfstate',
        roleArn: 'arn:aws:iam::123456789012:role/role',
      })
      expect(Object.keys(config.dependencies ?? {})).toEqual([
        'app',
        'lambda',
        'network',
      ])
      expect(config.providers?.awsRegional?.[0]).toMatchObject({
        alias: 'shared_services',
        region: 'ap-southeast-1',
        defaultTags: true,
      })
      expect(config.raw).toBeDefined()
    })

    it('only declares dependencies when filtered or forced', () => {
      expect(foggComponentToConfig(base()).dependencies).toEqual({})
      const filtered = { ...base(), component_backends_filtered: true }
      expect(
        Object.keys(foggComponentToConfig(filtered).dependencies ?? {}),
      ).toHaveLength(3)
    })

    it('throws on unsupported generic providers', () => {
      const c = base()
      c.required_providers = {
        bar: { enabled: true, source: 'foo/bar', version: '1', config: {} },
      } as any
      expect(() => foggComponentToConfig(c)).toThrow(
        /Unsupported provider foo\/bar/,
      )
    })

    it('does not mutate the input when defaulting module_name', () => {
      const c = { ...base(), module_source: 'mod' }
      const config = foggComponentToConfig(c)
      expect(c.module_name).toBeUndefined()
      expect((config.raw as Component).module_name).toBe('main')
      expect(config.mainModule).toEqual({ source: 'mod', name: 'main' })
    })
  })
})
