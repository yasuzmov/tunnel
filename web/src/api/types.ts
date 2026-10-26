export type AdminRole = 'super_admin' | 'admin'

/** 左侧菜单权限键，与后端 service/menu.go 保持一致。 */
export type MenuKey = 'dashboard' | 'tunnels' | 'domains' | 'logs' | 'admins'

export interface AdminUser {
  id: number
  username: string
  role: AdminRole
  menus: MenuKey[]
}

/** 管理员列表项，包含被授权的域名与可见菜单。 */
export interface AdminAccount {
  id: number
  username: string
  role: AdminRole
  created_at: string
  domain_ids: number[]
  domain_names: string[]
  menus: MenuKey[]
}

/** 统一域名池条目，包含当前绑定情况。 */
export interface ManagedDomain {
  id: number
  domain: string
  remark: string
  created_at: string
  bound_tunnel_id: number
  bound_tunnel_name: string
  https_enabled: boolean
  cert_status: string
  nginx_status: string
}

/** 当前管理员可绑定且尚未被占用的域名。 */
export interface AvailableDomain {
  id: number
  domain: string
  remark: string
}

export interface TunnelDomain {
  id: number
  tunnel_id: number
  domain: string
  https_enabled: boolean
  cert_status: string
  cert_expire_at?: string
  nginx_status: string
  last_error: string
  created_at: string
}

export interface Tunnel {
  id: number
  name: string
  remark: string
  key: string
  log_enabled: boolean
  last_online_at?: string
  /** 创建者管理员 ID；0 表示升级前的历史记录（无归属，仅超级管理员可见）。 */
  created_by: number
  /** 创建者登录名，仅超级管理员查询时返回。 */
  creator_name?: string
  created_at: string
  domains?: TunnelDomain[]
}

export interface TunnelDetail {
  tunnel: Tunnel
  connect_url: string
}

export interface ClientSession {
  id: number
  tunnel_id: number
  client_ip: string
  hostname: string
  os: string
  version: string
  connected_at: string
  disconnected_at?: string
  last_heartbeat_at: string
  disconnect_reason: string
  request_count: number
  bytes_in: number
  bytes_out: number
  online: boolean
}

export interface RequestLog {
  id: number
  request_id: string
  tunnel_id: number
  session_id?: number
  domain: string
  method: string
  scheme: string
  host: string
  path: string
  query_string: string
  query_params: string
  form_params: string
  remote_ip: string
  user_agent: string
  referer: string
  request_headers: string
  response_headers: string
  status: number
  request_size: number
  response_size: number
  received_at: string
  forwarded_at?: string
  first_byte_at?: string
  completed_at?: string
  duration_ms: number
  upstream_ms: number
  local_target: string
  matched_rule: string
  error_message: string
  is_websocket: boolean
  ws_message_count: number
  created_at: string
  body?: RequestLogBody
}

export interface RequestLogBody {
  request_body?: string
  request_truncated: boolean
  request_content_type: string
  response_body?: string
  response_truncated: boolean
  response_content_type: string
}

export interface DashboardStats {
  online_sessions: number
  tunnel_count: number
  today_requests: number
  today_errors: number
  error_rate: number
}
