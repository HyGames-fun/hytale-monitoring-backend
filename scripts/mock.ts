import { PrismaClient } from '../generated/prisma/client'
import { PrismaPg } from '@prisma/adapter-pg'

const adapter = new PrismaPg({
  connectionString: process.env.DATABASE_URL
})

const prisma = new PrismaClient({ adapter })

async function main() {
  await prisma.server.createMany({
    data: [
      {
        domain: 'go.hyfable.com',
        description:
          'Survival / PvE server focused on classic gameplay and exploration',
        name: 'Hyfable Survival',
        nameId: 'hyfable-survival',
        tags: ['SURVIVAL', 'PVE', 'VANILLA'],
        region: 'EUROPE'
      },
      {
        ip: '8.8.8.8:53',
        domain: 'play.hyrivals.gg',
        description: 'PvP competitive server with arenas and combat modes',
        name: 'HyRivals PvP',
        nameId: 'hyrivals-pvp',
        poster:
          'https://d1elhbmy4ij4xv.cloudfront.net/69_banner_307a3af79a.png',
        tags: ['PVP', 'SURVIVAL'],
        region: 'NORTHAMERICA'
      },
      {
        ip: '1.1.1.1:53',
        domain: 'play.elitehytale.com',
        description: 'PvP server with active community and battle systems',
        name: 'EliteHytale PvP',
        nameId: 'elitehytale-pvp',
        tags: ['PVP', 'MINIGAMES'],
        region: 'EUROPE'
      },
      {
        domain: 'play.valtale.com',
        description:
          'Adventure PvP/Survival with economy and leaderboard systems',
        name: 'Apex Hytale',
        nameId: 'apex-hytale',
        tags: ['PVP', 'SURVIVAL', 'RPG'],
        region: 'NORTHAMERICA'
      },
      {
        domain: 'play.moncube.eu',
        description: 'Un monde Aventure unique t’attend sur Moncube !',
        name: 'Moncube',
        nameId: 'moncube',
        poster: 'https://zupimages.net/up/26/04/u05d.png',
        tags: ['PVP', 'SURVIVAL'],
        region: 'EUROPE'
      },
      {
        ip: '1.1.1.1:81',
        domain: 'play.gohytale.net:15662',
        description: 'Anarchy / Survival server with PvE and PvP elements',
        name: 'GoHytale Anarchy',
        nameId: 'gohytale-anarchy',
        tags: ['PVE', 'PVP', 'SURVIVAL'],
        region: 'EUROPE'
      },
      {
        domain: 'play.everfall.fun',
        description: 'Long-term survival server with RPG mechanics',
        name: 'Everfall Survival',
        nameId: 'everfall-survival',
        tags: ['SURVIVAL', 'RPG', 'PVE'],
        region: 'EUROPE'
      }
    ]
  })
}

main()
  .catch((e) => {
    console.error(e)
    process.exit(1)
  })
  .finally(async () => {
    await prisma.$disconnect()
  })