CREATE TABLE IF NOT EXISTS users
(
    "id"       uuid PRIMARY KEY,
    email      varchar(50) NOT NULL UNIQUE,
    password   varchar(50) NOT NULL,
    role_id    uuid REFERENCES roles (id),
    first_name varchar(50) NOT NULL,
    last_name  varchar(50) NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS roles
(
    "id"  uuid PRIMARY KEY,
    title varchar(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS assistants
(
    id         uuid PRIMARY KEY,
    user_id    uuid  REFERENCES users (id),
    first_name varchar(50) NOT NULL,
    last_name  varchar(50) NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS doctors
(
    id         uuid PRIMARY KEY,
    user_id    uuid REFERENCES users (id),
    first_name varchar(50) NOT NULL,
    last_name  varchar(50) NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS clients
(
    id                uuid PRIMARY KEY,
    user_id           uuid REFERENCES users (id),
    phone_number      varchar(15) NOT NULL,
    first_appointment boolean NOT NULL,
    created_at        timestamp NOT NULL,
    updated_at        timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS appointments
(
    id           uuid PRIMARY KEY,
    doctor_id    uuid REFERENCES doctors (id),
    client_id    uuid REFERENCES clients (id),
    title        varchar(256) NOT NULL,
    content      text NOT NULL,
    status_id    uuid REFERENCES statuses (id),
    scheduled_at timestamp NOT NULL,
    created_at   timestamp NOT NULL,
    updated_at   timestamp NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_status_id ON appointments(status_id);

CREATE INDEX IF NOT EXISTS idx_doctor_id ON appointments(doctor_id);

CREATE INDEX IF NOT EXISTS idx_client_id ON appointments(client_id);

CREATE TABLE IF NOT EXISTS statuses
(
    id          uuid PRIMARY KEY,
    status_name varchar(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS medical_reports
(
    id             uuid PRIMARY KEY,
    diagnosis      text NOT NULL,
    recommendations text NOT NULL,
    appointment_id uuid REFERENCES appointments (id),
    attachment_id  uuid REFERENCES attachments (id),
    created_at     timestamp NOT NULL,
    updated_at     timestamp NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_appointment_id ON medical_reports(appointment_id);

CREATE TABLE IF NOT EXISTS attachments
(
    id              uuid PRIMARY KEY ,
    file_name       varchar(50) NOT NULL,
    file_url        varchar(100) NOT NULL,
    attachment_size varchar(6) NULL,
    attached_by_id  uuid REFERENCES doctors (id),
    attached_at     timestamp NOT NULL,
    updated_at      timestamp NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_attached_by_id ON attachments(attached_by_id);

CREATE TABLE IF NOT EXISTS reminders
(
    id             uuid PRIMARY KEY,
    appointment_id uuid REFERENCES appointments (id),
    client_id      uuid REFERENCES clients (id),
    content        varchar(1024) NOT NULL,
    read           boolean NOT NULL,
    created_at     timestamp NOT NULL,
    updated_at     timestamp NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_appointment_id ON reminders(appointment_id);

CREATE INDEX IF NOT EXISTS idx_client_id ON reminders(client_id);

CREATE TABLE IF NOT EXISTS reminder_settings
(
    id       uuid PRIMARY KEY,
    interval varchar(100) NOT NULL
);
