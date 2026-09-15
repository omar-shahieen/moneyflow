import { initContract } from "@ts-rest/core";
import { ZSubscriptionResponse, ZErrorResponse } from "@moneyflow/zod";
import { getSecurityMetadata } from "@/utils.js";

const c = initContract();

export const subscriptionContract = c.router({
  get: {
    summary: "Get current subscription",
    description: "Get the current user's subscription. Ensures a free subscription exists if none present.",
    path: "/api/v1/subscription",
    method: "GET",
    responses: {
      200: ZSubscriptionResponse,
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
});
