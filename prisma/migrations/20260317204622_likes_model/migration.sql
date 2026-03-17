/*
  Warnings:

  - You are about to drop the column `likes` on the `server` table. All the data in the column will be lost.
  - You are about to drop the `_user_liked_servers` table. If the table is not empty, all the data it contains will be lost.
  - A unique constraint covering the columns `[user_id]` on the table `server` will be added. If there are existing duplicate values, this will fail.

*/
-- DropForeignKey
ALTER TABLE "_user_liked_servers" DROP CONSTRAINT "_user_liked_servers_A_fkey";

-- DropForeignKey
ALTER TABLE "_user_liked_servers" DROP CONSTRAINT "_user_liked_servers_B_fkey";

-- AlterTable
ALTER TABLE "server" DROP COLUMN "likes",
ADD COLUMN     "likes_quantity" INTEGER NOT NULL DEFAULT 0;

-- DropTable
DROP TABLE "_user_liked_servers";

-- CreateTable
CREATE TABLE "Like" (
    "userId" INTEGER NOT NULL,
    "serverId" INTEGER NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "Like_pkey" PRIMARY KEY ("userId","serverId")
);

-- CreateIndex
CREATE UNIQUE INDEX "server_user_id_key" ON "server"("user_id");

-- AddForeignKey
ALTER TABLE "Like" ADD CONSTRAINT "Like_userId_fkey" FOREIGN KEY ("userId") REFERENCES "user"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Like" ADD CONSTRAINT "Like_serverId_fkey" FOREIGN KEY ("serverId") REFERENCES "server"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
