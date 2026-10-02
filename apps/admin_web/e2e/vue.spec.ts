import { test, expect } from '@playwright/test'

test('visits the app root url', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('h1')).toHaveText('平台状态')
  await expect(page).toHaveTitle('如此萌屋管理端')
})

test('shows the invalid verification code instead of a generic error', async ({ page }) => {
  await page.route('**/api/v1/admin/auth/login', async (route) => {
    await route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify({
        schema_version: '1.0.0',
        request_id: 'test-login',
        data: {
          status: 'mfa_required',
          challenge_token: 'test-challenge',
          account: {
            id: 'test-admin',
            email: 'admin@sprout.local',
            display_name: '管理员',
            role: 'admin',
            status: 'active',
          },
        },
        error: null,
      }),
    })
  })
  await page.route('**/api/v1/admin/auth/mfa', async (route) => {
    await route.fulfill({
      status: 401,
      contentType: 'application/json',
      body: JSON.stringify({
        schema_version: '1.0.0',
        request_id: 'test-mfa',
        data: null,
        error: {
          code: 'invalid_mfa_code',
          message: '验证码不正确，请重新输入',
          retryable: false,
        },
      }),
    })
  })

  await page.goto('/login')
  await page.getByLabel('管理员邮箱').fill('admin@sprout.local')
  await page.getByLabel('密码').fill('test-password')
  await page.getByRole('button', { name: '登录管理后台' }).click()
  await page.getByLabel('6 位验证码').fill('123456')
  await page.getByRole('button', { name: '验证并登录' }).click()

  await expect(page.getByText('验证码不正确，请重新输入')).toBeVisible()
  await expect(page.getByText('操作没有完成，请稍后重试')).toHaveCount(0)
})
