# Output and Privacy

Data commands such as accounts, codes, orders, charts, and `market tick` emit JSON on success.
`config set`, `config show`, and `config path` emit text. Parse JSON only for JSON-producing commands
and summarize only the fields the user needs.

## Privacy Rules

Never print, save, or commit:

- Kiwoom App Key
- Kiwoom Secret Key
- Issued access token
- Raw `~/.stock/config`
- Raw `~/.stock/token.json`

When checking configuration, use:

```bash
stock config show
stock config path
```

`config show` masks credentials.

## Sanitized Evidence

For verification or reports, prefer sanitized evidence:

- Command exit status.
- Top-level JSON keys.
- Row counts.
- Selected user-requested field values.
- Whether expected fields are present.

Avoid leaving full account, order, or chart output on disk. If temporary files are necessary, use a
private temp file and remove it after extracting evidence.

## Interpreting Live Order Results

Create output usually contains `order_id` and `trading_venue`. It does not prove that an order fully
filled.

Cancel output contains a new `order_id`, the `base_original_order_id`, and `cancelled_quantity`.

Use `stock orders list` after a live order only when the user asks or when it is necessary to answer
status questions.
