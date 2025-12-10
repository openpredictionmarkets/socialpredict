import { z } from "zod";

export const LoginResponseSchema = z.object({
  token: z.string(),
  username: z.string(),
  usertype: z.string(),
  mustChangePassword: z.boolean(),
});

export type LoginResponseParsed = z.infer<typeof LoginResponseSchema>;