import { describe, expect, it } from 'vitest';
import { resolveUploadErrorMessage } from './uploadError';

describe('resolveUploadErrorMessage', () => {
  it('prefers plain text response body', () => {
    expect(resolveUploadErrorMessage({
      response: {
        data: 'put s3 object failed: AccessDenied: Access Denied',
      },
    }, '上传失败')).toBe('put s3 object failed: AccessDenied: Access Denied');
  });

  it('falls back to structured message fields', () => {
    expect(resolveUploadErrorMessage({
      response: {
        data: {
          message: 'bucket policy denied this action',
        },
      },
    }, '上传失败')).toBe('bucket policy denied this action');
  });

  it('ignores generic axios status text and uses fallback', () => {
    expect(resolveUploadErrorMessage({
      message: 'Request failed with status code 500',
    }, '上传失败')).toBe('上传失败');
  });
});
