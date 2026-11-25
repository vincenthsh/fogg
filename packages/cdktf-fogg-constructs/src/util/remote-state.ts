import type { RemoteStateSchema } from '@tcons/component'
import {
  type InferOutputSchema,
  fromLegacySchema,
} from '@tcons/component/legacy'

/**
 * A `remoteState()` schema: a zod object or a v5 string schema
 * (`{ vpc_id: 'string', subnets: 'list' }`).
 */
export type FoggRemoteStateSchema<T = any> =
  | RemoteStateSchema
  | InferOutputSchema<T>

function isZodSchema(schema: unknown): schema is RemoteStateSchema {
  return typeof schema === 'object' && schema !== null && '_zod' in schema
}

/**
 * Normalize a `remoteState()` schema to zod: v5 string schemas are converted
 * with `fromLegacySchema` (rendered expressions are unchanged).
 */
export function toRemoteStateSchema(
  schema?: FoggRemoteStateSchema,
): RemoteStateSchema | undefined {
  if (schema === undefined || schema === null) return undefined
  if (isZodSchema(schema)) return schema
  return fromLegacySchema(schema)
}
