import { ElMessage } from 'element-plus'

// 优先使用异步剪贴板 API（仅 HTTPS / localhost 可用）；
// 不可用或被拒绝时（如 http 访问、嵌入式浏览器）回退到 execCommand，保证复制可用。
function fallbackCopy(text: string): boolean {
  const el = document.createElement('textarea')
  el.value = text
  el.setAttribute('readonly', '')
  el.style.position = 'fixed'
  el.style.top = '0'
  el.style.left = '0'
  el.style.opacity = '0'
  document.body.appendChild(el)
  el.focus()
  el.select()
  el.setSelectionRange(0, text.length)
  let done = false
  try {
    done = document.execCommand('copy')
  } catch {
    done = false
  }
  document.body.removeChild(el)
  return done
}

export async function copyText(text: string, ok = '已复制') {
  let done = false
  if (navigator.clipboard && window.isSecureContext) {
    try {
      await navigator.clipboard.writeText(text)
      done = true
    } catch {
      done = false
    }
  }
  if (!done) done = fallbackCopy(text)
  if (done) ElMessage.success(ok)
  else ElMessage.error('复制失败，请手动选中内容复制')
}
