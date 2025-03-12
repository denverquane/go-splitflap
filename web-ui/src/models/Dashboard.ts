import { z } from 'zod';
import { LocationSchema, SizeSchema } from './LocationSize';

// Routine types
export const RoutineType = z.enum(['TEXT', 'CLOCK', 'TIMER', 'SLOWTEXT', 'WEATHER']);

// Base schemas for routine interfaces
export const TextRoutineSchema = z.object({
  text: z.string(),
});

export const ClockRoutineSchema = z.object({
  format: z.string(),
});

export const TimerRoutineSchema = z.object({
  countdown: z.number(),
  format: z.string(),
});

export const SlowTextSchema = z.object({
  text: z.string(),
  delay: z.number(),
});

export const WeatherRoutineSchema = z.object({
  location: z.string(),
  api_key: z.string(),
});

// Create a discriminated union for routines
export const RoutineSchema = z.object({
  type: RoutineType,
  location: LocationSchema,
  size: SizeSchema,
  routine: z.union([
    TextRoutineSchema,
    ClockRoutineSchema,
    TimerRoutineSchema,
    SlowTextSchema,
    WeatherRoutineSchema,
  ]),
});

// Main Dashboard schema
export const DashboardSchema = z.object({
  routines: z.array(RoutineSchema),
});

// TypeScript types derived from the Zod schemas
export type Location = z.infer<typeof LocationSchema>;
export type Size = z.infer<typeof SizeSchema>;
export type TextRoutine = z.infer<typeof TextRoutineSchema>;
export type ClockRoutine = z.infer<typeof ClockRoutineSchema>;
export type TimerRoutine = z.infer<typeof TimerRoutineSchema>;
export type SlowText = z.infer<typeof SlowTextSchema>;
export type WeatherRoutine = z.infer<typeof WeatherRoutineSchema>;
export type Routine = z.infer<typeof RoutineSchema>;
export type Dashboard = z.infer<typeof DashboardSchema>;