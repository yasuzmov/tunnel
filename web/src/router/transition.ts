/**
 * 路由切换方向计算：根据 meta.depth（层级）与 meta.order（同级菜单顺序）
 * 决定新页从右侧滑入（前进）还是从左侧滑入（后退），让动画与用户的空间感一致。
 */
import { ref } from 'vue'
import type { RouteLocationNormalized } from 'vue-router'

export type SlideName = 'slide-left' | 'slide-right'

/** 当前应使用的过渡名，由 MainLayout 的 <transition :name> 读取。 */
export const slideName = ref<SlideName>('slide-left')

/**
 * 计算并写入过渡方向。
 * @param to 目标路由
 * @param from 来源路由
 * 副作用：修改 slideName。
 */
export function updateSlideDirection(to: RouteLocationNormalized, from: RouteLocationNormalized) {
  const toDepth = to.meta.depth ?? 1
  const fromDepth = from.meta.depth ?? 1
  if (toDepth !== fromDepth) {
    slideName.value = toDepth > fromDepth ? 'slide-left' : 'slide-right'
    return
  }
  const toOrder = to.meta.order ?? 0
  const fromOrder = from.meta.order ?? 0
  slideName.value = toOrder >= fromOrder ? 'slide-left' : 'slide-right'
}
