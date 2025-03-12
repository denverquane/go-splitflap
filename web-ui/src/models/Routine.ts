import { z } from 'zod';

// Parameter schema for routine parameters
export const ParameterSchema = z.object({
  name: z.string(),
  description: z.string(),
  field: z.string(),
  type: z.string(),
});

export type Parameter = z.infer<typeof ParameterSchema>;

// Use a more flexible record for config - only validate it's an object
export const RoutineConfigSchema = z.record(z.string(), z.any());

// Size schema for routine size constraints
export const RoutineSizeSchema = z.object({
  width: z.number(),
  height: z.number(),
});

// Schema for a full routine info object
export const RoutineInfoSchema = z.object({
  parameters: z.array(ParameterSchema),
  config: z.any(), // Accept any config object
  min_size: RoutineSizeSchema.optional(),
  max_size: RoutineSizeSchema.optional(),
});

// Schema for the full API response
export const RoutinesResponseSchema = z.record(z.string(), RoutineInfoSchema);

// Types derived from schemas
export type RoutineConfig = Record<string, any>;
export type RoutineInfo = z.infer<typeof RoutineInfoSchema>;
export type RoutinesResponse = z.infer<typeof RoutinesResponseSchema>;