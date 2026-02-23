-- CreateEnum
CREATE TYPE "Region" AS ENUM ('CIS', 'EUROPE', 'ASIA', 'AFRICA', 'NORTHAMERICA', 'SOUTHAMERICA');

-- CreateEnum
CREATE TYPE "Tag" AS ENUM ('PVP', 'PVE', 'RP', 'RPG', 'VANILLA', 'SURVIVAL', 'MINIGAMES');

-- CreateTable
CREATE TABLE "user" (
    "id" SERIAL NOT NULL,
    "name" TEXT NOT NULL,
    "email" TEXT,
    "avatar" TEXT,
    "password" TEXT,
    "discord_id" TEXT,

    CONSTRAINT "user_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "server" (
    "id" SERIAL NOT NULL,
    "domain" TEXT NOT NULL,
    "real_ip" TEXT,
    "description" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "name_id" TEXT NOT NULL,
    "tags" "Tag"[],
    "region" "Region" NOT NULL,
    "poster" TEXT,
    "likes" INTEGER NOT NULL DEFAULT 0,
    "user_id" INTEGER,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "server_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "_user_liked_servers" (
    "A" INTEGER NOT NULL,
    "B" INTEGER NOT NULL,

    CONSTRAINT "_user_liked_servers_AB_pkey" PRIMARY KEY ("A","B")
);

-- CreateIndex
CREATE UNIQUE INDEX "user_email_key" ON "user"("email");

-- CreateIndex
CREATE UNIQUE INDEX "user_discord_id_key" ON "user"("discord_id");

-- CreateIndex
CREATE UNIQUE INDEX "server_domain_key" ON "server"("domain");

-- CreateIndex
CREATE UNIQUE INDEX "server_real_ip_key" ON "server"("real_ip");

-- CreateIndex
CREATE UNIQUE INDEX "server_name_id_key" ON "server"("name_id");

-- CreateIndex
CREATE INDEX "_user_liked_servers_B_index" ON "_user_liked_servers"("B");

-- AddForeignKey
ALTER TABLE "server" ADD CONSTRAINT "server_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "user"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "_user_liked_servers" ADD CONSTRAINT "_user_liked_servers_A_fkey" FOREIGN KEY ("A") REFERENCES "server"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "_user_liked_servers" ADD CONSTRAINT "_user_liked_servers_B_fkey" FOREIGN KEY ("B") REFERENCES "user"("id") ON DELETE CASCADE ON UPDATE CASCADE;
