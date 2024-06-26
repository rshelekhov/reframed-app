CREATE TABLE IF NOT EXISTS lists
(
    id         character varying PRIMARY KEY,
    title      character varying NOT NULL,
    user_id    character varying NOT NULL,
    is_default boolean NOT NULL DEFAULT false,
    created_at timestamp WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at timestamp WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at timestamp WITH TIME ZONE DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_list_user_id ON lists (user_id);

CREATE TABLE IF NOT EXISTS tasks
(
    id          character varying PRIMARY KEY,
    title       character varying NOT NULL,
    description character varying,
    start_date  timestamp WITH TIME ZONE,
    deadline    timestamp WITH TIME ZONE,
    start_time  timestamp WITH TIME ZONE,
    end_time    timestamp WITH TIME ZONE,
    status_id   int NOT NULL,
    list_id     character varying NOT NULL,
    heading_id  character varying NOT NULL,
    user_id     character varying NOT NULL,
    created_at  timestamp WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at  timestamp WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at  timestamp WITH TIME ZONE DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_task_user_id ON tasks(user_id);

CREATE TABLE IF NOT EXISTS statuses
(
    id    int PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
    title character varying NOT NULL
);

CREATE TABLE IF NOT EXISTS headings
(
    id         character varying PRIMARY KEY,
    title      character varying NOT NULL,
    list_id    character varying NOT NULL,
    user_id    character varying NOT NULL,
    is_default boolean NOT NULL DEFAULT false,
    created_at timestamp WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at timestamp WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at timestamp WITH TIME ZONE DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS tags
(
    id         character varying PRIMARY KEY,
    title      character varying NOT NULL,
    user_id    character varying NOT NULL,
    created_at timestamp WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at timestamp WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at timestamp WITH TIME ZONE DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS tasks_tags
(
    task_id character varying NOT NULL,
    tag_id  character varying NOT NULL,
    CONSTRAINT tasks_tags_pkey PRIMARY KEY (task_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_tag_user_id ON tags(user_id);

CREATE TABLE IF NOT EXISTS reminders
(
    id         character varying PRIMARY KEY,
    content    character varying NOT NULL,
    read       boolean NOT NULL,
    task_id    character varying NOT NULL,
    user_id    character varying NOT NULL,
    created_at timestamp WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at timestamp WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at timestamp WITH TIME ZONE DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_remind_task_id ON reminders(task_id);

CREATE TABLE IF NOT EXISTS reminder_settings
(
    id       int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    interval character varying NOT NULL
);

ALTER TABLE tasks ADD FOREIGN KEY (status_id) REFERENCES statuses (id);
ALTER TABLE tasks ADD FOREIGN KEY (list_id) REFERENCES lists (id);
ALTER TABLE tasks ADD FOREIGN KEY (heading_id) REFERENCES headings(id);
ALTER TABLE headings ADD FOREIGN KEY (list_id) REFERENCES lists(id);
ALTER TABLE reminders ADD FOREIGN KEY (task_id) REFERENCES tasks(id);
