# ---------------------------------------------------
# مرحله ۱: کش و نصب پکیج‌ها
# ---------------------------------------------------
FROM node:24-alpine AS deps
WORKDIR /app

COPY package*.json ./
RUN npm ci

# ---------------------------------------------------
# مرحله ۲: کامپایل و بیلد پروژه
# ---------------------------------------------------
FROM node:24-alpine AS builder
WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY . .

ENV NODE_ENV=production
RUN npm run build

# ---------------------------------------------------
# مرحله ۳: کانتینر نهایی و مینیمال پروداکشن
# ---------------------------------------------------
FROM node:24-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3000
ENV HOST=0.0.0.0

# ساخت کاربر غیر ریشه برای امنیت در سطح Enterprise
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nuxtuser

# فقط پوشه خروجی نیترو منتقل می‌شود (فاقد سورس‌کد و node_modules حجیم)
COPY --from=builder --chown=nuxtuser:nodejs /app/.output ./.output

USER nuxtuser

EXPOSE 3000

# اجرای مستقیم سرور سبک Nitro
CMD ["node", ".output/server/index.mjs"]
