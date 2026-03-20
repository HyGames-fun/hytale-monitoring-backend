import { Injectable } from '@nestjs/common'
import { PrismaService } from './prisma/prisma.service'

@Injectable()
export class AppService {
  constructor(private readonly prisma: PrismaService) {}

  healthCheck() {
    void this.prisma.$executeRaw`SELECT 1`
  }
}
