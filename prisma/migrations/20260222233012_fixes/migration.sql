/*
  Warnings:

  - Added the required column `likes` to the `server` table without a default value. This is not possible if the table is not empty.

*/
-- AlterTable
ALTER TABLE "server" ADD COLUMN     "likes" INTEGER NOT NULL,
ADD COLUMN     "poster" TEXT;

-- AlterTable
ALTER TABLE "user" ALTER COLUMN "email" DROP NOT NULL;
