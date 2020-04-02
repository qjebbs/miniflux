// Code generated by go generate; DO NOT EDIT.

package database // import "miniflux.app/database"

var SqlMap = map[string]string{
	"schema_version_1": `create table schema_version (
    version text not null
);

create table users (
    id serial not null,
    username text not null unique,
    password text,
    is_admin bool default 'f',
    language text default 'en_US',
    timezone text default 'UTC',
    theme text default 'default',
    last_login_at timestamp with time zone,
    primary key (id)
);

create table sessions (
    id serial not null,
    user_id int not null,
    token text not null unique,
    created_at timestamp with time zone default now(),
    user_agent text,
    ip text,
    primary key (id),
    unique (user_id, token),
    foreign key (user_id) references users(id) on delete cascade
);

create table categories (
    id serial not null,
    user_id int not null,
    title text not null,
    primary key (id),
    unique (user_id, title),
    foreign key (user_id) references users(id) on delete cascade
);

create table feeds (
    id bigserial not null,
    user_id int not null,
    category_id int not null,
    title text not null,
    feed_url text not null,
    site_url text not null,
    checked_at timestamp with time zone default now(),
    etag_header text default '',
    last_modified_header text default '',
    parsing_error_msg text default '',
    parsing_error_count int default 0,
    primary key (id),
    unique (user_id, feed_url),
    foreign key (user_id) references users(id) on delete cascade,
    foreign key (category_id) references categories(id) on delete cascade
);

create type entry_status as enum('unread', 'read', 'removed');

create table entries (
    id bigserial not null,
    user_id int not null,
    feed_id bigint not null,
    hash text not null,
    published_at timestamp with time zone not null,
    title text not null,
    url text not null,
    author text,
    content text,
    status entry_status default 'unread',
    primary key (id),
    unique (feed_id, hash),
    foreign key (user_id) references users(id) on delete cascade,
    foreign key (feed_id) references feeds(id) on delete cascade
);

create index entries_feed_idx on entries using btree(feed_id);

create table enclosures (
    id bigserial not null,
    user_id int not null,
    entry_id bigint not null,
    url text not null,
    size int default 0,
    mime_type text default '',
    primary key (id),
    foreign key (user_id) references users(id) on delete cascade,
    foreign key (entry_id) references entries(id) on delete cascade
);

create table icons (
    id bigserial not null,
    hash text not null unique,
    mime_type text not null,
    content bytea not null,
    primary key (id)
);

create table feed_icons (
    feed_id bigint not null,
    icon_id bigint not null,
    primary key(feed_id, icon_id),
    foreign key (feed_id) references feeds(id) on delete cascade,
    foreign key (icon_id) references icons(id) on delete cascade
);
`,
	"schema_version_10": `drop table tokens;

create table sessions (
    id text not null,
    data jsonb not null,
    created_at timestamp with time zone not null default now(),
    primary key(id)
);`,
	"schema_version_11": `alter table integrations add column wallabag_enabled bool default 'f';
alter table integrations add column wallabag_url text default '';
alter table integrations add column wallabag_client_id text default '';
alter table integrations add column wallabag_client_secret text default '';
alter table integrations add column wallabag_username text default '';
alter table integrations add column wallabag_password text default '';`,
	"schema_version_12": `alter table entries add column starred bool default 'f';`,
	"schema_version_13": `create index entries_user_status_idx on entries(user_id, status);
create index feeds_user_category_idx on feeds(user_id, category_id);
`,
	"schema_version_14": `alter table integrations add column nunux_keeper_enabled bool default 'f';
alter table integrations add column nunux_keeper_url text default '';
alter table integrations add column nunux_keeper_api_key text default '';`,
	"schema_version_15": `alter table enclosures alter column size set data type bigint;`,
	"schema_version_16": `alter table entries add column comments_url text default '';`,
	"schema_version_17": `alter table integrations add column pocket_enabled bool default 'f';
alter table integrations add column pocket_access_token text default '';
alter table integrations add column pocket_consumer_key text default '';
`,
	"schema_version_18": `alter table user_sessions alter column ip set data type inet using ip::inet;`,
	"schema_version_19": `alter table feeds add column username text default '';
alter table feeds add column password text default '';`,
	"schema_version_2": `create extension if not exists hstore;
alter table users add column extra hstore;
create index users_extra_idx on users using gin(extra);
`,
	"schema_version_20": `alter table entries add column document_vectors tsvector;
update entries set document_vectors = setweight(to_tsvector(coalesce(title, '')), 'A') || setweight(to_tsvector(coalesce(content, '')), 'B');
create index document_vectors_idx on entries using gin(document_vectors);`,
	"schema_version_21": `alter table feeds add column user_agent text default '';`,
	"schema_version_22": `update entries set document_vectors = setweight(to_tsvector(substring(coalesce(title, '') for 1000000)), 'A') || setweight(to_tsvector(substring(coalesce(content, '') for 1000000)), 'B');`,
	"schema_version_23": `alter table users add column keyboard_shortcuts boolean default 't';`,
	"schema_version_24": `alter table feeds add column disabled boolean default 'f';`,
	"schema_version_25": `ALTER TABLE users ALTER COLUMN theme SET DEFAULT 'light_serif';
UPDATE users SET theme='light_serif' WHERE theme='default';
UPDATE users SET theme='light_sans_serif' WHERE theme='sansserif';
UPDATE users SET theme='dark_serif' WHERE theme='black';
`,
	"schema_version_26": `alter table entries add column changed_at timestamp with time zone;
update entries set changed_at = published_at;
alter table entries alter column changed_at set not null;
`,
	"schema_version_27": `create table api_keys (
    id serial not null,
    user_id int not null references users(id) on delete cascade,
    token text not null unique,
    description text not null,
    last_used_at timestamp with time zone,
    created_at timestamp with time zone default now(),
    primary key(id),
    unique (user_id, description)
);
`,
	"schema_version_28": `alter table entries add column share_code text not null default '';
create unique index entries_share_code_idx on entries using btree(share_code) where share_code <> '';
`,
	"schema_version_3": `create table tokens (
    id text not null,
    value text not null,
    created_at timestamp with time zone not null default now(),
    primary key(id, value)
);`,
	"schema_version_4": `create type entry_sorting_direction as enum('asc', 'desc');
alter table users add column entry_direction entry_sorting_direction default 'asc';
`,
	"schema_version_5": `create table integrations (
    user_id int not null,
    pinboard_enabled bool default 'f',
    pinboard_token text default '',
    pinboard_tags text default 'miniflux',
    pinboard_mark_as_unread bool default 'f',
    instapaper_enabled bool default 'f',
    instapaper_username text default '',
    instapaper_password text default '',
    fever_enabled bool default 'f',
    fever_username text default '',
    fever_password text default '',
    fever_token text default '',
    primary key(user_id)
)
`,
	"schema_version_6": `alter table feeds add column scraper_rules text default '';
`,
	"schema_version_7": `alter table feeds add column rewrite_rules text default '';
`,
	"schema_version_8": `alter table feeds add column crawler boolean default 'f';
`,
	"schema_version_9": `alter table sessions rename to user_sessions;`,
}

var SqlMapChecksums = map[string]string{
	"schema_version_1":  "00b2fa9e945565625c93ef9d4242a8b6583dc3cd7edf38d2fc95c0f3f7b926ae",
	"schema_version_10": "8faf15ddeff7c8cc305e66218face11ed92b97df2bdc2d0d7944d61441656795",
	"schema_version_11": "dc5bbc302e01e425b49c48ddcd8e29e3ab2bb8e73a6cd1858a6ba9fbec0b5243",
	"schema_version_12": "a95abab6cdf64811fc744abd37457e2928939d999c5ef00d2bdd9398e16f32fb",
	"schema_version_13": "9073fae1e796936f4a43a8120ebdb4218442fe7d346ace6387556a357c2d7edf",
	"schema_version_14": "4622e42c4a5a88b6fe1e61f3d367b295968f7260ab5b96481760775ba9f9e1fe",
	"schema_version_15": "13ff91462bdf4cda5a94a4c7a09f757761b0f2c32b4be713ba4786a4837750e4",
	"schema_version_16": "9d006faca62fd7ab787f64aef0e0a5933d142466ec4cab0e096bb920d2797e34",
	"schema_version_17": "b9f15d6217275fedcf6d948dd85ebe978b869bf37f42a86fd5b50a51919fa0e1",
	"schema_version_18": "c0ec24847612c7f2dc326cf735baffba79391a56aedd73292371a39f38724a71",
	"schema_version_19": "a83f77b41cc213d282805a5b518f15abbf96331599119f0ef4aca4be037add7b",
	"schema_version_2":  "e8e9ff32478df04fcddad10a34cba2e8bb1e67e7977b5bd6cdc4c31ec94282b4",
	"schema_version_20": "44790faf5806cccc9b785faee4f852554ce4dc7d67f2281548f2004902e857fd",
	"schema_version_21": "77da01ee38918ff4fe33985fbb20ed3276a717a7584c2ca9ebcf4d4ab6cb6910",
	"schema_version_22": "51ed5fbcae9877e57274511f0ef8c61d254ebd78dfbcbc043a2acd30f4c93ca3",
	"schema_version_23": "cb3512d328436447f114e305048c0daa8af7505cfe5eab02778b0de1156081b2",
	"schema_version_24": "1224754c5b9c6b4038599852bbe72656d21b09cb018d3970bd7c00f0019845bf",
	"schema_version_25": "5262d2d4c88d637b6603a1fcd4f68ad257bd59bd1adf89c58a18ee87b12050d7",
	"schema_version_26": "64f14add40691f18f514ac0eed10cd9b19c83a35e5c3d8e0bce667e0ceca9094",
	"schema_version_27": "4235396b37fd7f52ff6f7526416042bb1649701233e2d99f0bcd583834a0a967",
	"schema_version_28": "a64b5ba0b37fe3f209617b7d0e4dd05018d2b8362d2c9c528ba8cce19b77e326",
	"schema_version_3":  "a54745dbc1c51c000f74d4e5068f1e2f43e83309f023415b1749a47d5c1e0f12",
	"schema_version_4":  "216ea3a7d3e1704e40c797b5dc47456517c27dbb6ca98bf88812f4f63d74b5d9",
	"schema_version_5":  "46397e2f5f2c82116786127e9f6a403e975b14d2ca7b652a48cd1ba843e6a27c",
	"schema_version_6":  "9d05b4fb223f0e60efc716add5048b0ca9c37511cf2041721e20505d6d798ce4",
	"schema_version_7":  "33f298c9aa30d6de3ca28e1270df51c2884d7596f1283a75716e2aeb634cd05c",
	"schema_version_8":  "9922073fc4032d8922617ec6a6a07ae8d4817846c138760fb96cb5608ab83bfc",
	"schema_version_9":  "de5ba954752fe808a993feef5bf0c6f808e0a4ced5379de8bec8342678150892",
}
