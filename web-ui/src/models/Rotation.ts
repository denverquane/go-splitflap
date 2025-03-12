import { z } from 'zod';

// Schema for a single rotation entry (dashboard + duration)
export const RotationEntrySchema = z.object({
  name: z.string(),
  duration_secs: z.number(), // duration in seconds
});

// Schema for a rotation
export const RotationSchema = z.object({
  rotation: z.array(RotationEntrySchema),
});

// Schema for a map of rotations
export const RotationsSchema = z.record(z.string(), RotationSchema);

// TypeScript types derived from Zod schemas
export type RotationEntry = z.infer<typeof RotationEntrySchema>;
export type Rotation = z.infer<typeof RotationSchema>;
export type Rotations = z.infer<typeof RotationsSchema>;