/*
  Warnings:

  - You are about to drop the `_UserLikedServers` table. If the table is not empty, all the data it contains will be lost.

*/
-- DropForeignKey
ALTER TABLE "_UserLikedServers" DROP CONSTRAINT "_UserLikedServers_A_fkey";

-- DropForeignKey
ALTER TABLE "_UserLikedServers" DROP CONSTRAINT "_UserLikedServers_B_fkey";

-- DropTable
DROP TABLE "_UserLikedServers";

-- CreateTable
CREATE TABLE "_user_liked_servers" (
    "A" INTEGER NOT NULL,
    "B" INTEGER NOT NULL,

    CONSTRAINT "_user_liked_servers_AB_pkey" PRIMARY KEY ("A","B")
);

-- CreateIndex
CREATE INDEX "_user_liked_servers_B_index" ON "_user_liked_servers"("B");

-- AddForeignKey
ALTER TABLE "_user_liked_servers" ADD CONSTRAINT "_user_liked_servers_A_fkey" FOREIGN KEY ("A") REFERENCES "server"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "_user_liked_servers" ADD CONSTRAINT "_user_liked_servers_B_fkey" FOREIGN KEY ("B") REFERENCES "user"("id") ON DELETE CASCADE ON UPDATE CASCADE;
