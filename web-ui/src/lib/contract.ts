import { initContract } from "@ts-rest/core";
import { z } from "zod";
import { SizeSchema } from "@/models/LocationSize";
import { DashboardSchema } from "@/models/Dashboard";

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
    getTranslations: {
      method: "GET",
      path: "/display/translations",
      responses: {
        200: z.record(z.string(), z.string()),
        404: z.string(),
        500: z.string(),
      },
    },
    updateTranslations: {
      method: "POST",
      path: "/display/translations",
      body: z.record(z.string(), z.string()),
      responses: {
        200: z.object({ status: z.literal("ok") }),
        400: z.string(),
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
  },
});