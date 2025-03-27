-- SQLite
DROP TABLE IF EXISTS "comment_likes";
DROP TABLE IF EXISTS "post_likes";
DROP TABLE IF EXISTS "comments";
DROP TABLE IF EXISTS "post_categories";
DROP TABLE IF EXISTS "posts";
DROP TABLE IF EXISTS "categories";
DROP TABLE IF EXISTS "sessions";
DROP TABLE IF EXISTS "messages";
DROP TABLE IF EXISTS "chats";
DROP TABLE IF EXISTS "users";

CREATE TABLE "categories" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT,
  "name" TEXT NOT NULL,
  "status" TEXT NOT NULL CHECK ("status" IN ('enable', 'disable', 'delete')) DEFAULT 'enable',
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "created_by" INTEGER NOT NULL,
  "updated_at" DATETIME,
  "updated_by" INTEGER,
  FOREIGN KEY (created_by) REFERENCES "users" ("id"),
  FOREIGN KEY (updated_by) REFERENCES "users" ("id")
);

CREATE TABLE "users" (
  "id" INTEGER PRIMARY KEY,
  "uuid" TEXT NOT NULL UNIQUE,
  "type" TEXT NOT NULL CHECK ("type" IN ('admin', 'normal_user', 'test_user')) DEFAULT 'normal_user',
	"username" TEXT UNIQUE NOT NULL,
	"age" INTEGER NOT NULL,
	"gender" TEXT NOT NULL,
	"firstname" TEXT NOT NULL,
	"lastname" TEXT NOT NULL,
	"email" TEXT UNIQUE NOT NULL,
	"password" TEXT NOT NULL,
  "status" TEXT NOT NULL CHECK ("status" IN ('enable', 'disable', 'delete')) DEFAULT 'enable',
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" DATETIME,
  "updated_by" INTEGER,
  FOREIGN KEY (updated_by) REFERENCES "users" ("id")
);


CREATE TABLE "chats" (
  "id" INTEGER PRIMARY KEY,
  "uuid" TEXT NOT NULL UNIQUE,
  "user_id_1" INTEGER NOT NULL,
  "user_id_2" INTEGER NOT NULL,
  "status" TEXT NOT NULL CHECK ("status" IN ('enable', 'disable', 'delete')) DEFAULT 'enable',
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" DATETIME,
  "updated_by" INTEGER,
  FOREIGN KEY (user_id_1) REFERENCES "users" ("id"),
  FOREIGN KEY (user_id_2) REFERENCES "users" ("id"),
  FOREIGN KEY (updated_by) REFERENCES "users" ("id"),
  CONSTRAINT unique_chat UNIQUE (user_id_1, user_id_2) 
);

CREATE TABLE "messages" (
  "id" INTEGER PRIMARY KEY,
  "chat_id" INTEGER NOT NULL,
  "user_id_from" INTEGER NOT NULL,
  "content" TEXT NOT NULL,
  "status" TEXT NOT NULL CHECK ("status" IN ('enable', 'disable', 'delete')) DEFAULT 'enable',
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" DATETIME,
  FOREIGN KEY (chat_id) REFERENCES "chats" ("id"),
  FOREIGN KEY (user_id_from) REFERENCES "users" ("id")
);


CREATE TABLE "posts" (
  "id" INTEGER PRIMARY KEY,
  "uuid" TEXT NOT NULL UNIQUE,
  "title" TEXT NOT NULL,
  "description" TEXT NOT NULL,
  "status" TEXT NOT NULL CHECK ("status" IN ('enable', 'disable', 'delete')) DEFAULT 'enable',
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "user_id" INTEGER NOT NULL,
  "updated_at" DATETIME,
  "updated_by" INTEGER,
  FOREIGN KEY (user_id) REFERENCES "users" ("id"),
  FOREIGN KEY (updated_by) REFERENCES "users" ("id")
);

CREATE TABLE "post_likes" (
  "id" INTEGER PRIMARY KEY,
  "type" TEXT NOT NULL CHECK ("type" IN ('like', 'dislike')),
  "post_id" INTEGER NOT NULL,
  "user_id" INTEGER NOT NULL,
  "status" TEXT NOT NULL CHECK ("status" IN ('enable', 'delete')) DEFAULT 'enable',
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" DATETIME,
  "updated_by" INTEGER,
  FOREIGN KEY (user_id) REFERENCES "users" ("id"),
  FOREIGN KEY (post_id) REFERENCES "posts" ("id"),
  FOREIGN KEY (updated_by) REFERENCES "users" ("id")
);

CREATE TABLE "post_categories" (
  "id" INTEGER PRIMARY KEY,
  "post_id" INTEGER NOT NULL,
  "category_id" INTEGER NOT NULL,
  "status" TEXT NOT NULL CHECK ("status" IN ('enable', 'disable', 'delete')) DEFAULT 'enable',
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "created_by" INTEGER NOT NULL,
  "updated_at" DATETIME,
  "updated_by" INTEGER,
  FOREIGN KEY (created_by) REFERENCES "users" ("id"),
  FOREIGN KEY (updated_by) REFERENCES "users" ("id"),
  FOREIGN KEY (post_id) REFERENCES "posts" ("id"),
  FOREIGN KEY (category_id) REFERENCES "categories" ("id")
);

CREATE TABLE "comments" (
  "id" INTEGER PRIMARY KEY,
  "post_id" INTEGER DEFAULT NULL,
  "comment_id" INTEGER DEFAULT NULL,
  "description" TEXT NOT NULL,
  "user_id" INTEGER NOT NULL,
  "status" TEXT NOT NULL CHECK ("status" IN ('enable', 'disable', 'delete')) DEFAULT 'enable',
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" DATETIME,
  "updated_by" INTEGER,
  FOREIGN KEY (user_id) REFERENCES "users" ("id"),
  FOREIGN KEY (updated_by) REFERENCES "users" ("id"),
  FOREIGN KEY (post_id) REFERENCES "posts" ("id") ON DELETE CASCADE,
  FOREIGN KEY (comment_id) REFERENCES "comments" ("id") ON DELETE CASCADE,
  CHECK (
    (post_id IS NOT NULL AND comment_id IS NULL) OR
    (post_id IS NULL AND comment_id IS NOT NULL)
  )
);

CREATE TABLE "comment_likes" (
  "id" INTEGER PRIMARY KEY,
  "type" TEXT NOT NULL,
  "comment_id" INTEGER NOT NULL,
  "user_id" INTEGER NOT NULL,
  "status" TEXT NOT NULL CHECK ("status" IN ('enable', 'delete')) DEFAULT 'enable',
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" DATETIME,
  "updated_by" INTEGER,
  FOREIGN KEY (user_id) REFERENCES "users" ("id"),
  FOREIGN KEY (updated_by) REFERENCES "users" ("id"),
  FOREIGN KEY (comment_id) REFERENCES "comments" ("id")
);



CREATE TABLE "sessions" (
  "id" INTEGER PRIMARY KEY,
  "session_token" TEXT NOT NULL UNIQUE,
  "user_id" INTEGER NOT NULL,
  "expires_at" DATETIME NOT NULL,
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES "users" ("id")
);

INSERT INTO users(uuid, type, username, password, email, age, gender, firstname, lastname)
VALUES ('67921bdd-8458-800e-b9d4-065a43242cd3', 'admin', 'admin', '$2a$10$DN.v/NkfQjmPaTTz15x0E.u8l2R9.HnB12DpDVMdRPeQZDfMwovSa', 'admin@admin', 30, 'male', 'Admin', 'User');

INSERT INTO users(uuid, type, username, password, email, age, gender, firstname, lastname)
VALUES ('084d5c52-a72c-411c-a52f-6193f0614abe', 'normal_user', 'markus', '$2a$10$M7dekbtPuRH/hJ0qJr0mUeIL0KANj7IZ.cRPLz8e1PJtQ5A2aKjpO', 'ma@am.com', 42, 'male', 'Ma', 'Am');

INSERT INTO users(uuid, type, username, password, email, age, gender, firstname, lastname)
VALUES ('6952f31d-a07a-420b-a4dc-794271adec4f', 'normal_user', 'mahdi', '$2a$10$dmvXCLmw4QwthpYnKYzV9ue9zbgefnkoPdGxHlIc8YOhg3/LTUqw2', 'mh@kh.com', 22, 'male', 'Mh', 'Kh');


INSERT INTO categories (name, created_by)
VALUES ('art', 1), ('science', 1), ('news', 1);

INSERT INTO posts(uuid, title, description, user_id)
VALUES ('f9edb8d6-c739-4d6f-aaa4-9b298f2e1552', 'first post', 'this is first post of forum that is made by admin', 1);

INSERT INTO post_categories(post_id, category_id, created_by)
VALUES (1, 1, 1), (1, 2, 1);

INSERT INTO comments(post_id, description, user_id)
VALUES (1, 'this is first post comment that is made by admin', 1);

INSERT INTO post_likes(post_id, type, user_id)
VALUES (1, 'like', 1); 

INSERT INTO comment_likes(comment_id, type, user_id)
VALUES (1, 'like', 2); 