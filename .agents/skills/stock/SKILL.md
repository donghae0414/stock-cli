---
name: stock
description: >
  Use stock CLI for Kiwoom REST API and Korean stock workflows: account holdings, open orders, charts,
  tick-size checks, and live cash/credit stock order creation or cancellation.
  키움증권 stock CLI로 국내주식 보유종목, 주문 조회, 차트, 호가 단위, 현금/신용 매수·매도·취소를 처리합니다.
  Trigger this skill whenever the user mentions stock-cli, Kiwoom, 키움, 국내주식, 한국 주식,
  보유종목, 미체결, 주문, 매수, 매도, 취소, 호가 단위, or asks an agent to use the stock command.
metadata:
  version: v0.1.0
  author: stock-cli
license: Apache-2.0
---

# Stock Skill

Use the `stock` CLI binary for Kiwoom REST API and local Korean stock helper workflows.

## Language Behavior

Detect the user's language and respond accordingly:

- Korean user: respond in Korean and use Korean trading terms such as 보유종목, 미체결, 주문, 매수, 매도, 취소, 호가 단위.
- English user: respond in English and translate CLI fields into plain English.
- Mixed or ambiguous: follow the language of the most recent user message.

Load `references/glossary.md` when translating command names, trading terms, or JSON fields.

## Binary Selection

Prefer the repo-local built binary when working inside `/Users/dongwuk/apps/stock-cli`:

```bash
./bin/stock --help
```

If no repo-local binary exists, try:

```bash
stock --help
```

If neither works, load `references/setup.md`.

## Authentication

Private Kiwoom endpoints require credentials. Configure with:

```bash
stock config set
```

Credentials are stored in `~/.stock/config`. Issued access tokens are cached separately in
`~/.stock/token.json`. Never print, summarize, save, or commit real Kiwoom credentials or issued
tokens. Load `references/setup.md` for setup, credential priority, and token-cache policy.

Private commands include `accounts list`, `orders list`, `orders create`, `orders cancel`, and
`chart` commands. `market tick` is local-only and does not use credentials.

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
3. Show a concise risk summary containing side, funding type, stock code, order type, quantity,
   price behavior, trading venue, and any credit loan selection or loan date.
4. Ask the user to type exactly one word: `CONFIRM` or `확인`.
5. Execute the command only after the latest user message is exactly `CONFIRM` or exactly `확인`.

Important boundaries:

- A user's earlier approval, natural-language "yes", or broad instruction to proceed is not enough.
- Multi-word confirmations are not valid. Only the single-token response `CONFIRM` or `확인` is valid.
- Do not treat a confirmation embedded in the original order request as valid; show the final command
  first and wait for a new exact confirmation.
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
| Install, build, configure credentials, verify token/cache policy | `references/setup.md` |
| Holdings, available quantity, cash/credit detail | `references/accounts.md` |
| Open order list, buy, sell, cancel, cash vs credit order fields | `references/orders.md` |
| Day/week/minute Kiwoom chart queries | `references/charts.md` |
| Local Korean stock tick-size helper | `references/market.md` |
| JSON handling, privacy, and sanitized evidence | `references/output.md` |
| Korean/English terms and JSON field meanings | `references/glossary.md` |

## Common Workflow

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
stock accounts list
stock orders list --stock-code <six-digit-stock-code>
stock market tick --price <candidate-price>
```

Use `stock accounts list --credit-detail` when preparing credit sell orders that need loan-date
information.
