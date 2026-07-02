import { Message } from '@arco-design/web-vue'

export async function copyText(text: string, label?: string): Promise<void> {
  try {
    await navigator.clipboard.writeText(text)
    Message.success(label ? `已复制 ${label}` : '已复制')
  } catch {
    Message.error('复制失败')
  }
}
