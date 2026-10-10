┌──────────────────────┐
│        users         │
├──────────────────────┤
│ id       PK  uint    │◄──────────────┐
│ username UQ  varchar │               │
│ password     varchar │  (bcrypt 哈希) │
│ email    UQ  varchar │               │
│ created_at  datetime │               │
│ updated_at  datetime │               │
│ deleted_at  datetime │  (软删除)      │
└──────────────────────┘               │
▲                           │
│ user_id                   │ user_id
│                           │
┌──────────┴───────────┐   ┌───────────┴──────────┐
│        posts         │   │      comments        │
├──────────────────────┤   ├──────────────────────┤
│ id          PK       │◄──│ post_id     FK       │
│ title       varchar  │   │ id          PK       │
│ content     text     │   │ content     text     │
│ user_id     FK  ─────┘   │ created_at  datetime │
│ created_at  datetime │   └──────────────────────┘
│ updated_at  datetime │
│ deleted_at  datetime │
└──────────────────────┘

关系：一个 User 有多篇 Post（1:N）；一篇 Post 有多条 Comment（1:N）；
一条 Comment 属于一个 User 和一篇 Post。