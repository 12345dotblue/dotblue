export function resolveUploadErrorMessage(error: unknown, fallback: string): string {
  const messageFromResponse = firstNonEmpty(
    readString((error as any)?.response?.data),
    readString((error as any)?.response?.data?.message),
    readString((error as any)?.response?.data?.error),
  );
  if (messageFromResponse) {
    return messageFromResponse;
  }

  const genericMessage = readString((error as any)?.message);
  if (genericMessage && !/^Request failed with status code\s+\d+$/i.test(genericMessage)) {
    return genericMessage;
  }

  return fallback;
}

function readString(value: unknown): string {
  return typeof value === 'string' ? value.trim() : '';
}

function firstNonEmpty(...values: string[]): string {
  return values.find((value) => value.length > 0) || '';
}
