برای دبیان و این ساختار، بهترین و تمیزترین ابزار **`buf`** است که هم برای **Go** و هم برای **TypeScript (BFF)** خروجی تولید می‌کند.

مراحل نصب و اجرای دستورات را گام‌به‌گام در ادامه می‌آوریم:

---

### گام ۱: نصب ابزارهای مورد نیاز روی Debian

ابتدا ابزار `buf` و پلاگین‌های کامپایلر را نصب کنید:

```bash
sudo curl -sSL \
  "https://github.com/bufbuild/buf/releases/latest/download/buf-$(uname -s)-$(uname -m)" \
  -o /usr/local/bin/buf
sudo chmod +x /usr/local/bin/buf

# مطمئن شوید مسیر Go bin در PATH لینوکس شما هست:
export PATH="$PATH:$(go env GOPATH)/bin"
```

برای سمت تایپ‌اسکریپت (SvelteKit)، پلاگین‌های باینری به صورت محلی در `node_modules` نصب می‌شوند:

```bash
npm install -D @bufbuild/buf @bufbuild/protoc-gen-es
npm install @bufbuild/protovalidate
npm install @bufbuild/protobuf
```

برای سمت Go lang

```bash
go get buf.build/go/protovalidate
```

---

### گام ۲: تنظیم فایل `buf.gen.yaml`

در ریشه پروژه (یا داخل پوشه `proto/`)، فایل کانفیگ جنریتور را به شکل زیر بسازید که همزمان کدهای **Go** و **TypeScript** را تولید کند:

```yaml
version: v2

plugins:
  - remote: buf.build/protocolbuffers/go
    out: ../backend/gen/go
    opt:
      - paths=source_relative

  - remote: buf.build/bufbuild/es
    out: ../frontend/src/lib/gen
    opt:
      - target=ts
      - import_extension=none
    include_imports: true
```

> **نکته تحریم/شبکه:** در ایران اگر دسترسی به `buf.build` (Remote Plugins) به دلیل تحریم یا محدودیت شبکه تایم‌اوت خورد، می‌توانید از پلاگین‌های لوکال استفاده کنید:
>
> ```yaml
> version: v2
> plugins:
>   - local: protoc-gen-go
>     out: backend/gen/go
>     opt: paths=source_relative
>   - local: protoc-gen-es
>     out: frontend/src/lib/gen
>     opt: target=ts
> ```

---

### گام ۳: اجرای دستور تولید کد (Code Generation)

در ترمینال، دستور زیر را اجرا کنید:

```bash
# اگر buf.yaml و buf.gen.yaml در ریشه پروژه قرار دارند:
buf generate
```

اگر فایل‌های `.proto` داخل یک دایرکتوری خاص (مثلاً پوشه `proto`) قرار دارند:

```bash
buf generate proto
```

---

### گام کمکی: قرار دادن در `Makefile` یا `package.json`

برای اینکه هر بار نیاز به تایپ دستورات طولانی نداشته باشید، یک فایل `Makefile` در ریشه پروژه بسازید:

```makefile
.PHONY: proto

proto:
	@echo "Generating Protobuf stubs for Go and TypeScript..."
	buf generate
	@echo "Done!"
```

سپس تنها با دستور زیر در دبیان کدهایتان به‌روز می‌شوند:

```bash
make proto
```

پورت `1717` متعلق به دستور **SSH** است (احتمالاً با دستور `ssh -D 1717 ...` یک تونل داینامیک باز کرده‌اید).

تونل داینامیک SSH صرفاً یک **SOCKS5 Proxy** است و ترافیک HTTP عادی را مستقیماً متوجه نمی‌شود. به همین دلیل وقتی `HTTP_PROXY="http://..."` دادید، با خطای `Proxy CONNECT aborted` مواجه شدید.

علاوه بر این، در محیط‌های Go/Buf، ابزارها به متغیر استاندارد **`ALL_PROXY`** با فرمت SOCKS5 توجه می‌کنند.

---

### تست و فعال‌سازی روی همان SSH Tunnel:

۱. ابتدا با `curl` بررسی کنید که آیا تونل درست کار می‌کند یا خیر (از پروتکل `socks5h` استفاده کنید تا DNS هم از سرور SSH رد شود):

```bash
curl -x socks5h://127.0.0.1:1717 -I https://buf.build
```

باید خروجی با کد `200` یا `301` دریافت کنید.

---

۲. اگر تست بالا موفق بود، متغیرها را دقیقاً به شکل زیر در همان ترمینال ست کنید:

```bash
export ALL_PROXY="socks5h://127.0.0.1:1717"
export all_proxy="socks5h://127.0.0.1:1717"
export HTTP_PROXY="socks5h://127.0.0.1:1717"
export HTTPS_PROXY="socks5h://127.0.0.1:1717"
export http_proxy="socks5h://127.0.0.1:1717"
export https_proxy="socks5h://127.0.0.1:1717"
```

---

۳. حالا وارد پوشه proto شده و دستورات را اجرا کنید:

```bash
cd ~/Programs/Golang/sample-protobuf/proto
buf dep update
buf generate
```

این بار ارتباط با رجیستری `buf.build` از طریق سرور SSH شما عبور کرده و وابستگی‌ها و پلاگین‌ها بدون خطای 403 دانلود می‌شوند.
