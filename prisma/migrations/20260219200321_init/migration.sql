-- CreateEnum
CREATE TYPE "Region" AS ENUM ('CIS', 'EUROPE', 'ASIA', 'AFRICA', 'NORTHAMERICA', 'SOUTHAMERICA');

-- CreateEnum
CREATE TYPE "Tag" AS ENUM ('PVP', 'PVE', 'RP', 'RPG', 'VANILLA', 'SURVIVAL', 'MINIGAMES');

-- CreateTable
CREATE TABLE "Server" (
    "id" SERIAL NOT NULL,
    "ip" TEXT NOT NULL,
    "realIp" TEXT NOT NULL,
    "description" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "tags" "Tag"[],
    "region" "Region" NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "Server_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "Server_ip_key" ON "Server"("ip");

-- CreateIndex
CREATE UNIQUE INDEX "Server_realIp_key" ON "Server"("realIp");
