import { Injectable } from '@nestjs/common'
import { PrismaService } from './prisma/prisma.service'

@Injectable()
export class AppService {
  constructor(private readonly prisma: PrismaService) {}

  healthCheck() {
    return this.prisma.$queryRaw`SELECT 1`
  }
}
