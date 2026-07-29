import { createHmac } from "node:crypto";

const AUTH_DATE_MAX_AGE_SECONDS = 24 * 60 * 60;

export function verifyWebAppInitData(initData: string, botToken: string): boolean {
  const params = new URLSearchParams(initData);
  const hash = params.get("hash");
  const authDateValue = params.get("auth_date");

  if (!hash || !authDateValue) return false;

  const authDate = Number(authDateValue);
  if (!Number.isInteger(authDate) || authDate <= 0) return false;
  if (Math.floor(Date.now() / 1000) - authDate > AUTH_DATE_MAX_AGE_SECONDS) {
    return false;
  }

  const dataCheckString = Array.from(params.entries())
    .filter(([key]) => key !== "hash")
    .sort(([keyA], [keyB]) => keyA.localeCompare(keyB))
    .map(([key, value]) => `${key}=${value}`)
    .join("\n");

  const secretKey = createHmac("sha256", "WebAppData")
    .update(botToken)
    .digest();
  const computedHash = createHmac("sha256", secretKey)
    .update(dataCheckString)
    .digest("hex");

  return computedHash === hash;
}
