import { initContract } from "@ts-rest/core";
import { z } from "zod";
import { ZHealthResponse, ZErrorResponse } from "@moneyflow/zod";
import { getSecurityMetadata } from "@/utils.js";

const c = initContract();

export const healthContract = c.router({
  getHealth: {
    summary: "Get health",
    path: "/status",
    method: "GET",
    description: "Get health status. Returns 200 if healthy, 503 if unhealthy.",
    responses: {
      200: ZHealthResponse,
      503: ZHealthResponse,
    },
  },
});
