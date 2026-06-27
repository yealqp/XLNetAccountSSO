#!/usr/bin/env node

/**
 * .env 密钥 / 凭据生成脚本
 *
 * 生成内容:
 *   - MySQL 随机密码
 *   - RSA 2048-bit OIDC 签名私钥 (PKCS8 PEM)
 *   - OIDC Key ID (kid, 从公钥 SHA-256 派生)
 *   - WebAuthn RP ID
 *
 * 用法:
 *   node scripts/generate-env.mjs              # 输出到 stdout
 *   node scripts/generate-env.mjs > server/.env # 写入 server/.env
 *   node scripts/generate-env.mjs --root        # 输出根 .env（Docker 版）
 */

import crypto from 'node:crypto'

// ── 工具函数 ────────────────────────────────────────

/** 生成 N 字节的 hex 密码 */
function randomPassword(bytes = 16) {
  return crypto.randomBytes(bytes).toString('hex')
}

/** 生成随机 key ID (8 字节 base64url) */
function randomKeyID() {
  return crypto.randomBytes(8).toString('base64url')
}

/** 生成 RSA 2048 私钥并返回 { pem, kid } */
function generateOIDCKey() {
  // 不加 encoding → 返回 KeyObject，支持 .export()
  const { privateKey, publicKey } = crypto.generateKeyPairSync('rsa', {
    modulusLength: 2048,
  })

  // 私钥：PKCS8 PEM → 单行（\n 转义），适配 .env 等号后面的值
  const privatePEM = privateKey
    .export({ type: 'pkcs8', format: 'pem' })
    .trim()
    .replace(/\n/g, '\\n')

  // Key ID：公钥 DER → SHA-256[:16] → base64url（与 Go 端 deriveOIDCKeyID 一致）
  const der = publicKey.export({ type: 'spki', format: 'der' })
  const hash = crypto.createHash('sha256').update(der).digest()
  const kid = hash.subarray(0, 16).toString('base64url')

  return { pem: privatePEM, kid }
}

/** 将 WebBaseURL 解析为主机名作为 RP ID */
function rpIDFromURL(url) {
  try {
    return new URL(url).hostname
  } catch {
    return 'localhost'
  }
}

// ── 输出 ────────────────────────────────────────────

function generateServerEnv() {
  const dbPassword = randomPassword()
  const { pem: oidcPEM, kid: oidcKID } = generateOIDCKey()
  const rpID = rpIDFromURL('http://localhost:8080')
  const origins = 'http://localhost:5173,http://localhost:8080'

  return [
    '# ============================================',
    '# SSO Server - 自动生成配置',
    '# 生成时间: ' + new Date().toISOString(),
    '# ============================================',
    '',
    '# ── 服务端口 ──',
    'PORT=8080',
    '',
    '# ── 数据目录（图标等） ──',
    'ASSET_DIR=./data',
    '',
    '# ── 地址配置 ──',
    'SERVER_BASE_URL=http://localhost:8080',
    'WEB_BASE_URL=http://localhost:8080',
    'CORS_ORIGINS=' + origins,
    '',
    '# ── WebAuthn / Passkey ──',
    'WEBAUTHN_RP_ID=' + rpID,
    'WEBAUTHN_RP_ORIGINS=' + origins,
    '',
    '# ── OIDC ──',
    'OIDC_ISSUER=',
    'OIDC_KEY_ID=' + oidcKID,
    'OIDC_PRIVATE_KEY_PEM=' + oidcPEM,
    '',
    '# ── MySQL 数据库 ──',
    'DB_HOST=127.0.0.1',
    'DB_PORT=3306',
    'DB_USER=sso_app',
    'DB_PASSWORD=' + dbPassword,
    'DB_NAME=sso_platform',
    'DB_DSN=',
    '',
  ].join('\n')
}

function generateDockerEnv() {
  const dbPassword = randomPassword()
  const rootPassword = randomPassword()
  const { pem: oidcPEM, kid: oidcKID } = generateOIDCKey()
  const rpID = rpIDFromURL('http://localhost:8080')

  return [
    '# ============================================',
    '# SSO Docker - 自动生成配置',
    '# 生成时间: ' + new Date().toISOString(),
    '# ============================================',
    '',
    '# ── MySQL 数据库 ──',
    'SSO_MYSQL_DATABASE=sso_platform',
    'SSO_MYSQL_USER=sso_app',
    'SSO_MYSQL_PASSWORD=' + dbPassword,
    'SSO_MYSQL_ROOT_PASSWORD=' + rootPassword,
    '',
    '# ── CORS / 前端地址 ──',
    'SSO_CORS_ORIGINS=http://localhost:8080',
    'SSO_WEB_BASE_URL=http://localhost:8080',
    '',
    '# ── WebAuthn / Passkey ──',
    'SSO_WEBAUTHN_RP_ID=' + rpID,
    'SSO_WEBAUTHN_RP_ORIGINS=',
    '',
    '# ── OIDC ──',
    'SSO_OIDC_ISSUER=',
    'SSO_OIDC_KEY_ID=' + oidcKID,
    'SSO_OIDC_PRIVATE_KEY_PEM=' + oidcPEM,
    '',
  ].join('\n')
}

// ── 入口 ────────────────────────────────────────────

const args = process.argv.slice(2)

if (args.includes('--root') || args.includes('-r')) {
  process.stdout.write(generateDockerEnv())
} else {
  process.stdout.write(generateServerEnv())
}
