CREATE TABLE IF NOT EXISTS users
(
    id         uuid PRIMARY KEY,
    email      varchar(50) NOT NULL UNIQUE,
    password   varchar(50) NOT NULL,
    role_id    int NOT NULL,
    first_name varchar(50) NOT NULL,
    last_name  varchar(50) NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS roles
(
    id    int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    title varchar(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS assistants
(
    id         uuid PRIMARY KEY,
    user_id    uuid NOT NULL,
    first_name varchar(50) NOT NULL,
    last_name  varchar(50) NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS doctors
(
    id         uuid PRIMARY KEY,
    user_id    uuid NOT NULL,
    first_name varchar(50) NOT NULL,
    last_name  varchar(50) NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS clients
(
    id                uuid PRIMARY KEY,
    user_id           uuid NOT NULL,
    phone_number      varchar(15) NOT NULL,
    first_appointment boolean NOT NULL,
    created_at        timestamp NOT NULL,
    updated_at        timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS appointments
(
    id           uuid PRIMARY KEY,
    doctor_id    uuid NOT NULL,
    client_id    uuid NOT NULL,
    title        varchar(256) NOT NULL,
    content      text NOT NULL,
    status_id    int NOT NULL,
    scheduled_at timestamp NOT NULL,
    created_at   timestamp NOT NULL,
    updated_at   timestamp NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_status_id ON appointments(status_id);
CREATE INDEX IF NOT EXISTS idx_doctor_id ON appointments(doctor_id);
CREATE INDEX IF NOT EXISTS idx_client_id ON appointments(client_id);

CREATE TABLE IF NOT EXISTS statuses
(
    id          int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    status_name varchar(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS medical_reports
(
    id              uuid PRIMARY KEY,
    diagnosis       text NOT NULL,
    recommendations text NOT NULL,
    appointment_id  uuid NOT NULL,
    attachment_id   uuid NOT NULL,
    created_at      timestamp NOT NULL,
    updated_at      timestamp NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_appointment_id ON medical_reports(appointment_id);

CREATE TABLE IF NOT EXISTS attachments
(
    id              uuid PRIMARY KEY ,
    file_name       varchar(50) NOT NULL,
    file_url        varchar(100) NOT NULL,
    attachment_size varchar(6) NULL,
    attached_by_id  uuid NOT NULL,
    attached_at     timestamp NOT NULL,
    updated_at      timestamp NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_attached_by_id ON attachments(attached_by_id);

CREATE TABLE IF NOT EXISTS reminders
(
    id             uuid PRIMARY KEY,
    appointment_id uuid NOT NULL,
    client_id      uuid NOT NULL,
    content        varchar(1024) NOT NULL,
    read           boolean NOT NULL,
    created_at     timestamp NOT NULL,
    updated_at     timestamp NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_appointment_id ON reminders(appointment_id);
CREATE INDEX IF NOT EXISTS idx_client_id ON reminders(client_id);

CREATE TABLE IF NOT EXISTS reminder_settings
(
    id       int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    interval varchar(100) NOT NULL
);

ALTER TABLE users ADD FOREIGN KEY (role_id) REFERENCES roles (id);
ALTER TABLE assistants ADD FOREIGN KEY (user_id) REFERENCES users (id);
ALTER TABLE doctors ADD FOREIGN KEY (user_id) REFERENCES users (id);
ALTER TABLE clients ADD FOREIGN KEY (user_id) REFERENCES users (id);
ALTER TABLE appointments ADD FOREIGN KEY (doctor_id) REFERENCES doctors (id);
ALTER TABLE appointments ADD FOREIGN KEY (client_id) REFERENCES clients (id);
ALTER TABLE appointments ADD FOREIGN KEY (status_id) REFERENCES statuses (id);
ALTER TABLE medical_reports ADD FOREIGN KEY (appointment_id) REFERENCES appointments (id);
ALTER TABLE medical_reports ADD FOREIGN KEY (attachment_id) REFERENCES attachments (id);
ALTER TABLE attachments ADD FOREIGN KEY (attached_by_id) REFERENCES doctors (id);
ALTER TABLE reminders ADD FOREIGN KEY (appointment_id) REFERENCES appointments (id);
ALTER TABLE reminders ADD FOREIGN KEY (client_id) REFERENCES clients (id);
