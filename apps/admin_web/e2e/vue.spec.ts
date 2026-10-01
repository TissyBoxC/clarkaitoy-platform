import { test, expect } from '@playwright/test'

test('visits the app root url', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('h1')).toHaveText('平台状态')
  await expect(page).toHaveTitle('如此萌屋管理端')
})
