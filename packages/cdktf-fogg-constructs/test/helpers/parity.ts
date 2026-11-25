/**
 * Parity harness: runs a test suite against the published
 * @vincenthsh/cdktf-fogg-helpers 5.0.1 (`fogg-helpers-v5`) and this package
 * (v6), and asserts that every stack a v6 test constructs synthesizes
 * byte-identically to the stack the same v5 test constructed.
 *
 * Both versions read `.fogg-component.yaml` from the cwd, so `givenComponent`
 * writes the fixture into a per-test temp dir that the process chdir()s into
 * (vitest runs this package with the `forks` pool).
 *
 * Usage:
 *
 * ```ts
 * setupParity()
 * describe.each(IMPLS)('FoggStack [$name]', ({ FoggStack }) => { ... })
 * ```
 */
import * as fs from 'node:fs'
import * as os from 'node:os'
import * as path from 'node:path'
import { type TerraformStack, Testing } from 'cdktn'
import * as v5 from 'fogg-helpers-v5'
import * as yaml from 'js-yaml'
import { afterEach, beforeEach, expect } from 'vitest'
import * as v6 from '../../src'

export type ImplName = 'v5' | 'v6'

export interface Impl {
  readonly name: ImplName
  readonly FoggStack: typeof v6.FoggStack
  readonly FoggTerraStack: typeof v6.FoggTerraStack
}

type StackClass = new (...args: any[]) => TerraformStack

/** Stacks fully constructed by the current test, per implementation. */
const constructed: Array<{ impl: ImplName; stack: TerraformStack }> = []

function tracked<C extends StackClass>(impl: ImplName, Base: C): C {
  return class extends Base {
    constructor(...args: any[]) {
      super(...args)
      constructed.push({ impl, stack: this })
    }
  }
}

/**
 * v5.0.1 first, so v6 tests can compare against the recorded v5 synth.
 */
export const IMPLS: Impl[] = [
  {
    name: 'v5',
    FoggStack: tracked('v5', v5.FoggStack) as unknown as typeof v6.FoggStack,
    FoggTerraStack: tracked(
      'v5',
      v5.FoggTerraStack,
    ) as unknown as typeof v6.FoggTerraStack,
  },
  {
    name: 'v6',
    FoggStack: tracked('v6', v6.FoggStack),
    FoggTerraStack: tracked('v6', v6.FoggTerraStack),
  },
]

let tmpDir: string | undefined
let prevCwd: string | undefined

/**
 * Write `component` as `.fogg-component.yaml` into the test's working dir.
 */
export function givenComponent(component: v6.Component): void {
  if (!tmpDir) throw new Error('call setupParity() in the test file')
  fs.writeFileSync(
    path.join(tmpDir, '.fogg-component.yaml'),
    yaml.dump(component, { skipInvalid: true }),
  )
}

/**
 * Synthesize a stack the way the test `Template` helper does; errors are
 * captured so they are compared too.
 */
export function synthOrError(stack: TerraformStack): string {
  try {
    stack.prepareStack()
    return Testing.synth(stack)
  } catch (e) {
    return `ERROR: ${(e as Error).message}`
  }
}

/** Synth of every v5 stack, keyed by test path (without impl label). */
const v5Synths = new Map<string, string[]>()
let comparisons = 0

const IMPL_LABEL = / \['?(v[56])'?\]/

function testPath(task: { name: string; suite?: any }): string {
  const names: string[] = [task.name]
  for (let s = task.suite; s; s = s.suite) {
    if (s.name) names.unshift(s.name)
  }
  return names.join(' > ')
}

/**
 * Install the per-test temp cwd and the v5 / v6 synth comparison hooks.
 *
 * Inside a `describe.each(IMPLS)` suite, every v6 test must construct the
 * same number of stacks as its v5 run, each synthesizing to the same bytes.
 */
export function setupParity(): void {
  beforeEach(() => {
    constructed.length = 0
    prevCwd = process.cwd()
    tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'fogg-parity-'))
    process.chdir(tmpDir)
  })

  afterEach((ctx) => {
    try {
      const fullPath = testPath(ctx.task as any)
      const impl = IMPL_LABEL.exec(fullPath)?.[1]
      const key = fullPath.replace(IMPL_LABEL, '')
      const synths = constructed
        .splice(0)
        .map(({ stack }) => synthOrError(stack))
      if (impl === 'v5') {
        v5Synths.set(key, synths)
      } else if (impl === 'v6') {
        const expected = v5Synths.get(key)
        expect(expected, `no v5.0.1 run recorded for "${key}"`).toBeDefined()
        expect(synths.length, `stack count differs for "${key}"`).toBe(
          expected!.length,
        )
        for (const [i, synth] of synths.entries()) {
          comparisons++
          expect(synth, `v6 synth differs from v5.0.1 for "${key}"`).toBe(
            expected![i],
          )
        }
      }
    } finally {
      if (prevCwd) process.chdir(prevCwd)
      if (tmpDir) fs.rmSync(tmpDir, { recursive: true, force: true })
      tmpDir = undefined
    }
  })
}

/** Number of v6/v5 stack synth comparisons made so far. */
export function parityComparisons(): number {
  return comparisons
}
