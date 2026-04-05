export interface ScopeOption {
  key: string
  label: string
  description: string
}

export const scopeOptions: ScopeOption[] = [
  {
    key: 'openid',
    label: '身份认证',
    description: '启用 OIDC 登录并签发身份令牌。',
  },
  {
    key: 'profile',
    label: '基础资料',
    description: '读取账号基础信息。',
  },
  {
    key: 'email',
    label: '邮箱信息',
    description: '读取账号邮箱。',
  },
  {
    key: 'roles',
    label: '角色信息',
    description: '读取账号角色。',
  },
  {
    key: 'offline_access',
    label: '离线访问',
    description: '允许签发刷新令牌。',
  },
]

export const defaultScopeKeys = ['profile']

export const scopeOptionMap = new Map(scopeOptions.map(scope => [scope.key, scope]))
