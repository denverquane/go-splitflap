import { initContract } from "@ts-rest/core";
import { z } from "zod";
import { SizeSchema } from "@/models/LocationSize";
import { DashboardSchema } from "@/models/Dashboard";
import { RotationSchema, RotationsSchema } from "@/models/Rotation";

const c = initContract();

export const displayContract = c.router({
  display: {
    getSize: {
      method: "GET",
      path: "/display/size",
      responses: {
        200: SizeSchema,
        404: z.string(),
        500: z.string(),
      },
    },
    getDashboards: {
      method: "GET",
      path: "/dashboards",
      responses: {
        200: z.record(z.string(), DashboardSchema),
        404: z.string(),
        500: z.string(),
      },
    },
    activateDashboard: {
      method: "POST",
      path: "/dashboards/:name/activate",
      pathParams: z.object({
        name: z.string(),
      }),
      body: z.object({}), // Empty body - no body parameters required
      responses: {
        200: z.string(), // Returns the name of the activated dashboard
        400: z.string(),
        404: z.string(),
        500: z.string(),
      },
    },
    deleteDashboard: {
      method: "DELETE",
      path: "/dashboards/:name",
      pathParams: z.object({
        name: z.string(),
      }),
      responses: {
        200: z.string(), // Returns the name of the deleted dashboard
        400: z.string(),
        404: z.string(),
        500: z.string(),
      },
    },
    getRotations: {
      method: "GET",
      path: "/rotations",
      responses: {
        200: z.record(z.string(), RotationSchema),
        404: z.string(),
        500: z.string(),
      },
    },
    activateRotation: {
      method: "POST",
      path: "/rotations/:name/activate",
      pathParams: z.object({
        name: z.string(),
      }),
      body: z.object({}), // Empty body
      responses: {
        200: z.string(), // Returns the name of the activated rotation
        400: z.string(),
        404: z.string(),
        500: z.string(),
      },
    },
    deactivateRotation: {
      method: "POST",
      path: "/rotations/deactivate",
      body: z.object({}), // Empty body
      responses: {
        200: z.string(),
        400: z.string(),
        500: z.string(),
      },
    },
    createOrUpdateRotation: {
      method: "POST",
      path: "/rotations/:name",
      pathParams: z.object({
        name: z.string(),
      }),
      body: RotationSchema,
      responses: {
        200: z.string(), // Returns the name of the created/updated rotation
        400: z.string(),
        500: z.string(),
      },
    },
    deleteRotation: {
      method: "DELETE",
      path: "/rotations/:name",
      pathParams: z.object({
        name: z.string(),
      }),
      responses: {
        200: z.string(), // Returns the name of the deleted rotation
        400: z.string(),
        404: z.string(),
        500: z.string(),
      },
    },
  },
});