# Calculator

A precision calculator web application: a React/TypeScript single-page app talking to a Go REST API that evaluates arithmetic expressions. The frontend renders a calculator keypad with Unicode math operators (÷, ×, −) and validates expressions client-side against a formal grammar; the backend parses and evaluates the same grammar server-side and returns the result (or an RFC 9457 Problem Detail error).

```
┌─────────────────┐   POST /evaluations    ┌──────────────────┐
│  frontend (SPA)  │ ─────────────────────► │  backend (API)   │
│  React + Vite    │ ◄───────────────────── │  Go net/http     │
│  localhost:5173  │     JSON result        │  localhost:8080  │
└─────────────────┘                        └──────────────────┘
```

## Features

- **Unicode Math Operators**: Uses proper Unicode characters (÷, ×, −) for mathematical operations
- **Smart Expression Grammar**: Client-side validation prevents invalid expressions before reaching the backend
- **Responsive Design**: Adapts seamlessly from phones to landscape tablets with container-aware sizing
- **Keyboard Support**: Full keyboard input alongside touch controls
- **Optical Typography**: Carefully calibrated text positioning for glyph precision
- **Dark Mode**: System preference detection for light and dark themes
- **Error Recovery**: Expressions remain visible for correction when errors occur

## Project Layout

```
.
├── backend/    # Go REST API that parses and evaluates expressions
├── frontend/   # React SPA — the calculator UI
├── docs/adr/   # Architecture Decision Records
└── docker-compose.yaml
```

## Prerequisites

| Tool | Version | Needed for |
|------|---------|-------------|
| Go | 1.26+ | Running/building the backend (see `backend/go.mod`) |
| Node.js | 20.19+ (or 22.12+) — 24 recommended, see `frontend/.nvmrc` | Running/building the frontend |
| npm | bundled with Node | Frontend package management |
| Docker & Docker Compose | any recent version | Running both services together without installing Go/Node locally |

## Quick Start

### Option A — Docker Compose (both services, no local Go/Node install)

```bash
docker compose up --build
```

This builds and starts both containers:
- Backend API on `http://localhost:8080`
- Frontend (served by Nginx, which proxies `/evaluations` to the backend) on `http://localhost`

### Option B — Run locally for development

Run the backend first, then the frontend, in two terminals.

**1. Start the backend API:**

```bash
cd backend

# CALCULATOR_ALLOWED_ORIGINS is required — it lists browser origins allowed to call the API
CALCULATOR_ALLOWED_ORIGINS=http://localhost:5173 go run ./cmd/api
```

The API listens on `:8080` by default (override with `CALCULATOR_ADDR`, e.g. `CALCULATOR_ADDR=:9090`).

**2. Start the frontend dev server:**

```bash
cd frontend
npm install
npm run dev
```

The dev server runs on `http://localhost:5173` and proxies `/evaluations` requests to `http://127.0.0.1:8080`. Open `http://localhost:5173` in a browser to use the calculator.

### Building for production

```bash
cd frontend
npm run build      # outputs static assets to frontend/dist
npm run preview    # serve the production build locally
```

```bash
cd backend
go build -o api ./cmd/api
CALCULATOR_ALLOWED_ORIGINS=https://your-frontend-domain go run ./cmd/api
```

## API

The calculator communicates with a backend API using a simple REST interface.

### Evaluate Expression

**Endpoint:** `POST /evaluations`

**Request:**
- Content-Type: `text/plain; charset=utf-8`
- Body: The mathematical expression as plain text

```bash
curl -X POST http://localhost:8080/evaluations \
  -H "Content-Type: text/plain; charset=utf-8" \
  -d "7 + 3"
```

**Response (Success - 200):**
```json
{
  "data": {
    "result": 10
  }
}
```

**Response (Error - 4xx/5xx):**
Follows RFC 9457 Problem Detail specification:
```json
{
  "type": "https://example.com/errors/invalid-expression",
  "title": "Invalid Expression",
  "status": 400,
  "detail": "Unexpected token: @",
  "instance": "/evaluations"
}
```

### Expression Grammar

The backend accepts expressions following this grammar:

```
expression := number (operator number)* (percent)?
number     := digits ('.' digits)?
operator   := '+' | '−' | '×' | '÷'
percent    := '%'
```

**Key Rules:**
- Operators require operands on both sides
- Numbers may contain at most one decimal point
- Percent is postfix: `5%` calculates a percentage of the left operand
- Division by zero returns an error

**Example Expressions:**
- `42 + 8` → `50`
- `100 × 0.5` → `50`
- `50 − 15` → `35`
- `10 ÷ 2` → `5`
- `80 × 0.5%` → `0.4` (0.5% of 80)

## Configuration

The backend is configured entirely through `CALCULATOR_`-prefixed environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|--------------|
| `CALCULATOR_ALLOWED_ORIGINS` | Yes | — | Comma-separated list of browser origins allowed to call the API (CORS), e.g. `http://localhost:5173,https://example.com` |
| `CALCULATOR_ADDR` | No | `:8080` | Address the HTTP server listens on |

## Testing

### Backend

```bash
cd backend

# Run tests
go test ./...

# Generate coverage report
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out    # terminal summary
go tool cover -html=coverage.out    # opens an HTML report in the browser
```

### Frontend

```bash
cd frontend

# Run tests
npm run test

# Run tests with UI
npm run test:ui

# Generate coverage report (text summary + coverage/lcov-report/index.html)
npm run test:coverage
```

Coverage thresholds (80% lines/statements/functions/branches) are enforced in `vite.config.ts`; `npm run test:coverage` fails if they're not met.

## Linting

```bash
cd frontend
npm run lint
```

## Architecture

### Backend (`backend/`)

A Go `net/http` API with two internal packages. All of the interesting logic lives in `internal/expression`, which turns raw request text into a result in three steps: lex → parse → evaluate.

#### Lexer (`internal/expression/lexer.go`)

The lexer (`newLexer`) reads the request body rune-by-rune through a `bufio.Reader` and eagerly tokenizes the entire input up front — there's no separate scanning-on-demand step. It:

- Skips whitespace between tokens
- Recognizes only the exact Unicode operator runes in `operators.go` (`+`, `−`, `×`, `÷`, `%`) — ASCII lookalikes like `-` or `*` are rejected as `ErrUnrecognizedToken`, not silently accepted (see [ADR-002](docs/adr/0002-unicode-math-operators.md))
- Reads numbers strictly: a digit, then any run of digits with at most one `.`, and the `.` must be followed by another digit — `1.`, `.5`, and `1..2` all fail with `ErrInvalidNumber` rather than being coerced
- Reverses the resulting token slice once lexing completes, so the parser can treat it as a stack and pop tokens off the end in `O(1)` via `next()`/`peek()`

#### Parser (`internal/expression/expression.go`)

Parsing is a **Pratt parser** (precedence climbing / binding-power parser), implemented in `parseExpression`, that builds an `Expression` tree directly from the token stack — there's no separate AST-then-walk pass:

- Each operator has a **binding power** (`infixBindingPower`, `postfixBindingPower`) instead of a fixed grammar rule per precedence level: `+`/`−` bind at `1.0/1.1`, `×`/`÷` at `2.0/2.1`, and postfix `%` at `3.0` — higher binds tighter, so multiplication/division and percent naturally bind tighter than addition/subtraction without a separate grammar production for each level
- The right binding power of each infix operator is slightly higher than its left (`1.0` vs `1.1`, `2.0` vs `2.1`), which is what makes operators **left-associative** — `1 − 2 − 3` parses as `(1 − 2) − 3`, not `1 − (2 − 3)`
- `%` is postfix and has no right-hand side, so the parser special-cases `Operator.IsPostfix()` before it would otherwise look up an infix binding power
- A malformed sequence (an operator with no left operand, two atoms in a row, trailing operator, empty input) fails as `ErrUnexpectedToken` naming what was expected vs. what was found

#### Evaluator (`expression.go`: `Eval` methods)

Evaluation walks the tree built by the parser and is entirely separate from parsing — `New()` only parses; a caller must call `Eval()` to get a value. The tree has three node kinds, named by operand count rather than by operator, each implementing `Expression.Eval() (float32, error)`:

- `atomExpression`: a leaf holding a numeric literal — zero operands
- `unaryExpression`: one operand — currently only ever built for postfix `%`
- `binaryExpression`: exactly two operands — built for `+`, `−`, `×`, `÷`

- `binaryExpression.Eval` evaluates both operands depth-first, then applies the operator
- Percent is contextual: evaluated alone it's a hundredth (`unaryExpression.Eval`), but when it's the right operand of `+`/`−`, `binaryExpression.Eval` detects that and treats it as a **share of the left operand** instead — so `200 − 10%` is `180`, while `200 × 10%` is `20` (the plain hundredth) — see [ADR-002](docs/adr/0002-unicode-math-operators.md) and the package doc comment at the top of `expression.go` for the full rationale
- Division by a zero right operand returns `ErrDivisionByZero` rather than a float `Inf`
- Every `Expression` also implements `String()`, rendering itself as an S-expression (e.g. `1 + 2 × 3` → `(+ 1 (× 2 3))`), which makes operator precedence directly visible/testable rather than needing to reconstruct it from the tree in tests

All of these steps return one of the sentinel errors in `errors.go` (`ErrUnrecognizedToken`, `ErrUnexpectedToken`, `ErrInvalidNumber`, `ErrUnknownOperator`, `ErrDivisionByZero`, `ErrOperandCount`), which `httpx.go`'s `ClassifyError` maps to RFC 9457 Problem Details for the HTTP layer.

- `internal/httpx`: shared HTTP plumbing — CORS handling, JSON response helpers, and a `Kit` (logger + error classifier) injected into handlers

### Frontend (`frontend/src/`)

- **Expression Validation** (`lib/expression.ts`): maintains a grammar model to prevent invalid expressions from reaching the backend, avoiding unnecessary round-trips and giving immediate visual feedback
  - `append(expression, key)`: Add a character while validating grammar rules
  - `backspace(expression)`: Remove the last character
  - `isSubmittable(expression)`: Check if ready to send to backend
  - `forDisplay(expression)`: Format with thousand separators for display
- **State Management** (`components/Calculator.tsx`): the calculator operates in three phases:
  - **editing**: User is typing an expression
  - **result**: Backend returned a result; operator carries it forward, digit starts fresh
  - **pending**: Request in flight; input is ignored
- **Responsive Design**:
  - **Container Queries**: Key labels scale relative to keypad width, not viewport
  - **Dynamic Viewport Units**: `dvh` (dynamic viewport height) prevents layout shifts on mobile
  - **Custom Variant**: `@panel` applies bounded panel styles on wider screens
  - **Fit-Text Hook**: Display text shrinks to fit container width
- **Error Handling**: Errors follow RFC 9457 Problem Detail format and are displayed as dismissible toasts, preserving the expression for correction

## Design Decisions

See the [`docs/adr/`](docs/adr/) directory for Architecture Decision Records documenting key design choices:

- [ADR-001: Client-side Expression Grammar Validation](docs/adr/0001-client-side-grammar-validation.md)
- [ADR-002: Unicode Math Operators](docs/adr/0002-unicode-math-operators.md)
- [ADR-003: Three-Phase State Machine](docs/adr/0003-three-phase-state-machine.md)
- [ADR-004: Expression Grouping for Display](docs/adr/0004-expression-grouping-for-display.md)
- [ADR-005: Container Queries for Responsive Design](docs/adr/0005-container-queries-responsive-design.md)
- [ADR-006: Optical Adjustments for Typography](docs/adr/0006-optical-adjustments-typography.md)

## Technology Stack

**Backend:**
- **Go 1.26**: HTTP API, no third-party framework — standard library `net/http`
- **slog**: Structured JSON logging

**Frontend:**
- **React 19**: UI framework
- **TypeScript 6**: Type-safe development
- **Vite 8**: Build tooling
- **Tailwind CSS 4**: Utility-first styling
- **class-variance-authority**: Component variant management
- **Vitest**: Test framework
- **Sonner**: Toast notifications
- **Radix UI**: Accessible components

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go              # Entry point — server setup, env config, graceful shutdown
├── internal/
│   ├── expression/               # Grammar lexer/parser/evaluator + error classification
│   └── httpx/                    # CORS, JSON response helpers, shared handler Kit
└── go.mod

frontend/
├── src/
│   ├── components/
│   │   ├── Calculator.tsx       # Main calculator component
│   │   └── ui/                  # Shadcn UI components
│   ├── hooks/
│   │   └── use-fit-text.ts      # Responsive text sizing
│   ├── lib/
│   │   ├── api.ts               # API client
│   │   ├── expression.ts        # Grammar validation
│   │   └── utils.ts             # Utilities
│   ├── pages/
│   │   └── IndexPage.tsx        # Root page
│   ├── App.tsx                  # Root component
│   ├── main.tsx                 # Entry point
│   └── index.css                # Global styles
├── public/                       # Static assets
└── package.json
```

## Contributing

To report bugs or request features, please open an issue or pull request.

## License

See the LICENSE file for details.
