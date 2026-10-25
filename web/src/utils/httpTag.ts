/**
 * HTTP 方法 / 状态码到 Element Plus Tag 类型的映射，日志列表与详情共用，保证配色一致。
 */
export type TagType = 'primary' | 'success' | 'warning' | 'danger' | 'info'

const METHOD_TAGS: Record<string, TagType> = {
  GET: 'success',
  POST: 'primary',
  PUT: 'warning',
  PATCH: 'warning',
  DELETE: 'danger',
}

/**
 * @param method HTTP 方法，大小写不敏感
 * @returns 对应的 Tag 类型，未知方法返回 info
 */
export function methodTagType(method: string): TagType {
  return METHOD_TAGS[method?.toUpperCase()] ?? 'info'
}

/**
 * @param status HTTP 状态码；0 表示未拿到响应（转发失败或仍在进行）
 * @returns 2xx success、3xx info、4xx warning、5xx 与 0 为 danger
 */
export function statusTagType(status: number): TagType {
  if (!status || status >= 500) return 'danger'
  if (status >= 400) return 'warning'
  if (status >= 300) return 'info'
  return 'success'
}

/** 耗时着色阈值（毫秒）：<300 正常，<1000 偏慢，否则慢。 */
export function durationClass(ms: number): string {
  if (ms < 300) return 'fast'
  if (ms < 1000) return 'slow'
  return 'very-slow'
}
