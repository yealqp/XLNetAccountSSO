export interface ScopeOption {
  key: string
  label: string
  description: string
}

export const scopeOptions: ScopeOption[] = [
  {
    key: 'openid',
    label: 'OpenID',
    description: '启用 OIDC 登录并签发 id token。',
  },
  {
    key: 'profile',
    label: 'Profile',
    description: '读取账号基础信息。',
  },
  {
    key: 'email',
    label: 'Email',
    description: '读取账号邮箱。',
  },
  {
    key: 'roles',
    label: 'Roles',
    description: '读取账号角色。',
  },
  {
    key: 'offline_access',
    label: 'Offline Access',
    description: '允许签发 refresh token。',
  },
]

export const defaultScopeKeys = ['profile']

export const scopeOptionMap = new Map(scopeOptions.map(scope => [scope.key, scope]))
