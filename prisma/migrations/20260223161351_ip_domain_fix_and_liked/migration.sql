/*
  Warnings:

  - You are about to drop the column `ip` on the `server` table. All the data in the column will be lost.
  - A unique constraint covering the columns `[domain]` on the table `server` will be added. If there are existing duplicate values, this will fail.
  - Added the required column `domain` to the `server` table without a default value. This is not possible if the table is not empty.

*/
-- DropIndex
DROP INDEX "server_ip_key";

-- AlterTable
ALTER TABLE "server" DROP COLUMN "ip",
ADD COLUMN     "domain" TEXT NOT NULL;

-- CreateTable
CREATE TABLE "_UserLikedServers" (
    "A" INTEGER NOT NULL,
    "B" INTEGER NOT NULL,

    CONSTRAINT "_UserLikedServers_AB_pkey" PRIMARY KEY ("A","B")
);

-- CreateIndex
CREATE INDEX "_UserLikedServers_B_index" ON "_UserLikedServers"("B");

-- CreateIndex
CREATE UNIQUE INDEX "server_domain_key" ON "server"("domain");

-- AddForeignKey
ALTER TABLE "_UserLikedServers" ADD CONSTRAINT "_UserLikedServers_A_fkey" FOREIGN KEY ("A") REFERENCES "server"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "_UserLikedServers" ADD CONSTRAINT "_UserLikedServers_B_fkey" FOREIGN KEY ("B") REFERENCES "user"("id") ON DELETE CASCADE ON UPDATE CASCADE;
