import type {
  ExportedHandler,
  ForwardableEmailMessage,
  ExecutionContext,
} from "./cf-types";
import { handleEmail, type Env } from "./lib";

// Runtime wiring only. All behavior lives in handleEmail, which is unit-tested
// in lib.test.ts.
export default {
  async email(
    message: ForwardableEmailMessage,
    env: Env,
    _ctx: ExecutionContext
  ): Promise<void> {
    await handleEmail(message, env, fetch);
  },
} satisfies ExportedHandler<Env>;
