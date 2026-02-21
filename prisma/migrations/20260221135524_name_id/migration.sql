/*
  Warnings:

  - A unique constraint covering the columns `[name_id]` on the table `server` will be added. If there are existing duplicate values, this will fail.
  - Added the required column `name_id` to the `server` table without a default value. This is not possible if the table is not empty.

*/
-- AlterTable
ALTER TABLE "server" ADD COLUMN     "name_id" TEXT NOT NULL;

-- CreateIndex
CREATE UNIQUE INDEX "server_name_id_key" ON "server"("name_id");
