برای دبیان و این ساختار، بهترین و تمیزترین ابزار **`buf`** است که هم برای **Go** و هم برای **TypeScript (BFF)** خروجی تولید می‌کند.

مراحل نصب و اجرای دستورات را گام‌به‌گام در ادامه می‌آوریم:

---

### گام ۱: نصب ابزارهای مورد نیاز روی Debian

ابتدا ابزار `buf` و پلاگین‌های کامپایلر را نصب کنید:

```bash
# ۱. نصب باینری رسمی Buf
PREFIX="/usr/local"
VERSION="1.30.0" # یا جدیدترین نسخه
curl -sSL \
  "https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m)" \
  -o "${PREFIX}/bin/buf" && \
  chmod +x "${PREFIX}/bin/buf"

# ۲. نصب پلاگین‌های Go (با Go Toolchain)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/bufbuild/buf-plugin-protovalidate/cmd/protoc-gen-buf-validate@latest

# مطمئن شوید مسیر Go bin در PATH لینوکس شما هست:
export PATH="$PATH:$(go env GOPATH)/bin"
```

برای سمت تایپ‌اسکریپت (SvelteKit)، پلاگین‌های باینری به صورت محلی در `node_modules` نصب می‌شوند:

```bash
npm install -D @bufbuild/buf @bufbuild/protoc-gen-es
npm install @bufbuild/protobuf
```

---

### گام ۲: تنظیم فایل `buf.gen.yaml`

در ریشه پروژه (یا داخل پوشه `proto/`)، فایل کانفیگ جنریتور را به شکل زیر بسازید که همزمان کدهای **Go** و **TypeScript** را تولید کند:

```yaml
version: v2
plugins:
  # ۱. جنریت کدهای Go
  - remote: buf.build/protocolbuffers/go
    out: backend/gen/go
    opt:
      - paths=source_relative

  # ۲. جنریت کدهای اعتبارسنجی Go (protovalidate)
  - remote: buf.build/bufbuild/validate-go
    out: backend/gen/go
    opt:
      - paths=source_relative

  # ۳. جنریت کدهای TypeScript برای SvelteKit (BFF)
  - remote: buf.build/bufbuild/es
    out: frontend/src/lib/gen
    opt:
      - target=ts
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
