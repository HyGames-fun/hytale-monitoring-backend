import { Injectable } from '@nestjs/common'
import { ServerDto } from './server.dto'
import { PrismaService } from '../../prisma/prisma.service'

@Injectable()
export class ServerService {
  constructor(private readonly prisma: PrismaService) {}

  create(createServerDto: ServerDto) {
    return 'This action adds a new server'
  }

  findAll() {
    return `This action returns all server`
  }

  findOne(id: number) {
    return `This action returns a #${id} server`
  }

  remove(id: number) {
    return `This action removes a #${id} server`
  }
}
