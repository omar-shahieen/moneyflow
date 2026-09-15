import { initContract } from "@ts-rest/core";
import { z } from "zod";
import {
  ZUserAccount,
  ZUpdateUserRequest,
  ZErrorResponse,
} from "@moneyflow/zod";
import { getSecurityMetadata } from "@/utils.js";

const c = initContract();

export const userContract = c.router({
  getProfile: {
    summary: "Get user profile",
    description: "Get the authenticated user's profile.",
    path: "/api/v1/user",
    method: "GET",
    responses: {
      200: ZUserAccount,
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
  updateProfile: {
    summary: "Update user profile",
    description: "Update the authenticated user's profile.",
    path: "/api/v1/user",
    method: "PATCH",
    body: ZUpdateUserRequest,
    responses: {
      200: ZUserAccount,
      400: ZErrorResponse,
      401: ZErrorResponse,
    },
    metadata: getSecurityMetadata(),
  },
});
