import { ElMessage } from 'element-plus'

export async function copyText(text: string, ok = '已复制') {
  await navigator.clipboard.writeText(text)
  ElMessage.success(ok)
}
