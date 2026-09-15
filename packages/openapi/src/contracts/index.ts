import { initContract } from "@ts-rest/core";
import { healthContract } from "./health.js";
import { categoryContract } from "./category.js";
import { transactionContract } from "./transaction.js";
import { budgetContract } from "./budget.js";
import { subscriptionContract } from "./subscription.js";
import { recurringRuleContract } from "./recurring.js";
import { importContract } from "./import.js";
import { reportContract } from "./report.js";
import { userContract } from "./user.js";
import { notificationContract } from "./notification.js";

const c = initContract();

export const apiContract = c.router({
  Health: healthContract,
  Categories: categoryContract,
  Transactions: transactionContract,
  Budgets: budgetContract,
  Subscriptions: subscriptionContract,
  RecurringRules: recurringRuleContract,
  Imports: importContract,
  Reports: reportContract,
  User: userContract,
  Notifications: notificationContract,
});
