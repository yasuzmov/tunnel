import { reactive } from 'vue'

export function usePagination() {
  const pager = reactive({ page: 1, pageSize: 20, total: 0 })

  function onSizeChange(size: number) {
    pager.pageSize = size
    pager.page = 1
  }

  return { pager, onSizeChange }
}
