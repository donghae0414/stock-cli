---
name: kiwoom-domestic-stock
description: >
  Operate the stock CLI for Kiwoom domestic Korean equities: resolve stock names and codes, inspect
  holdings or open orders, query charts, check tick sizes, and prepare or execute cash/credit orders.
  키움증권 stock CLI로 국내 주식 종목코드 조회, 보유종목, 미체결, 차트, 호가 단위와
  현금/신용 주문 업무를 처리합니다. Use this skill when the user explicitly mentions stock-cli,
  stock command, Kiwoom, 키움, or 키움증권, or clearly asks to operate a Korean domestic-equity
  account/trading workflow. Do not use it for generic shopping orders or cancellations, general
  investing discussion, overseas/US stocks, crypto, or stock-cli repository development unless the
  request is to operate the CLI itself.
metadata:
  version: v0.1.0
  author: stock-cli
license: MIT
---

# Kiwoom Domestic Stock Skill

Use the `stock` CLI binary for Kiwoom REST API and domestic Korean stock helper workflows.

## Language Behavior

Detect the user's language and respond accordingly:

- Korean user: respond in Korean and use Korean trading terms such as 보유종목, 미체결, 주문, 매수, 매도, 취소, 호가 단위.
- English user: respond in English and translate CLI fields into plain English.
- Mixed or ambiguous: follow the language of the most recent user message.

Load `references/glossary.md` when translating command names, trading terms, or JSON fields.

## Setup and Authentication

If `stock` is not installed or credentials are not configured, load `references/setup.md` and follow
the steps there. Use the installed `stock` binary and start with:

```bash
stock --version
stock --help
```

Private Kiwoom endpoints require credentials configured by `stock config set`. Long-lived credentials
live in `~/.stock/config`; issued access tokens are cached separately in `~/.stock/token.json`. Never
print, summarize, save, or commit their values. Use `stock config show` only for masked configuration
evidence. `stock market tick` is local-only and does not load credentials.

Private commands include `accounts list`, `codes lookup`, `orders list`, `orders create`,
`orders cancel`, and `chart` commands. `market tick` is local-only and does not use credentials.

## Safety Rule - Live Trading Commands

The stock CLI order primitives submit live Kiwoom requests immediately. The CLI currently does not
provide a dry-run, quote confirmation, balance check, holding check, collateral check, price-limit
check, or built-in confirmation prompt. The agent using this skill is responsible for the safety
workflow.

This is a skill-layer safety contract for agents. It does not make the `stock` binary reject direct
unconfirmed calls. If the user asks for runtime-level enforcement, implement a CLI-side confirmation
or dry-run guard before claiming a hard binary-level guarantee.

Before executing any live trading command, stop and ask for explicit user confirmation.

Live trading commands:

- `stock orders create cash`
- `stock orders create credit`
- `stock orders cancel cash`
- `stock orders cancel credit`

Mandatory confirmation protocol:

1. Build the exact command, but do not execute it.
2. Show the exact command in a fenced `bash` block.
3. Show the matching risk summary:
   - For create: side, funding type, stock code and known name, order type, quantity, price behavior,
     trading venue, and any credit loan selection or loan date.
   - For cancel: funding type, stock code and known name, original order ID, requested cancel
     quantity or all remaining, and trading venue. Include the original order side and remaining
     quantity only when known from open-order evidence; do not invent them.
4. Ask the user to type exactly one word: `CONFIRM` (case-insensitive) or `확인`.
5. Execute the command only after the latest user message is a single token equal to `CONFIRM`
   ignoring ASCII case, or exactly `확인`.

Important boundaries:

- A user's earlier approval, natural-language "yes", or broad instruction to proceed is not enough.
- Multi-word confirmations are not valid. Only a single-token `CONFIRM` in any letter case, or the
  exact single-token response `확인`, is valid.
- Do not treat a confirmation embedded in the original order request as valid; show the final command
  first and wait for a fresh confirmation that matches the protocol above.
- If any command detail changes after confirmation, repeat the confirmation protocol.
- If the user asks to buy or sell but does not provide all required fields, collect the missing
  fields before showing the final command.
- Do not claim the binary itself enforces confirmation; this skill enforces the rule for agents that
  load and follow it.
- Use read-only commands such as `accounts list`, `orders list`, `chart`, and `market tick` without
  the live-trading confirmation gate, while still protecting credentials and tokens.

## Command References

Load only the reference needed for the user's task:

| Task | Reference |
| --- | --- |
| Install with npm, configure credentials, verify token/cache policy | `references/setup.md` |
| Resolve a Korean stock name to a six-digit code and handle candidates | `references/codes.md` |
| Holdings, available quantity, cash/credit detail | `references/accounts.md` |
| Open order list, buy, sell, cancel, cash vs credit order fields | `references/orders.md` |
| Day/week/minute Kiwoom chart queries | `references/charts.md` |
| Local Korean stock tick-size helper | `references/market.md` |
| JSON handling, privacy, and sanitized evidence | `references/output.md` |
| Korean/English terms and JSON field meanings | `references/glossary.md` |

## Common Workflow

When the user supplies a stock name but the command requires a six-digit code:

1. Load `references/codes.md` and run `stock codes lookup --name <name>` when credentials and the
   user's scope allow a private lookup.
2. Continue automatically only when that query has `status: "exact"` and exactly one candidate.
3. Ask the user to choose a candidate for `single_partial` or `ambiguous`. Stop on `not_found` or any
   error. Never infer a code for a live workflow from a partial match.

For read-only portfolio or market analysis:

1. Load the relevant reference.
2. Run the smallest command that answers the question.
3. Parse JSON output; summarize user-facing meaning without exposing secrets.

For live buy, sell, or cancel workflows:

1. Load `references/orders.md`.
2. Use read-only commands first when helpful, such as holdings, open orders, chart, or tick-size
   checks.
3. Prepare the exact live command.
4. Apply the mandatory confirmation protocol above.
5. After confirmed execution, summarize the normalized JSON result and any returned order id.

## First-Time Order Checklist

Before preparing a live buy or sell for an unfamiliar stock, prefer these read-only checks when
credentials and task scope allow:

```bash
stock codes lookup --name <stock-name>
stock accounts list
stock orders list --stock-code <six-digit-stock-code>
stock market tick --price <candidate-price>
```

Use `stock accounts list --credit-detail` when preparing credit sell orders that need loan-date
information.
