import { BadRequestException, ConflictException, Injectable, UnauthorizedException } from '@nestjs/common'
import { CreateServerDto } from './server.dto'
import { PrismaService } from '../../prisma/prisma.service'
import { User } from '../../../generated/prisma/client'

@Injectable()
export class ServerService {
  constructor(private readonly prisma: PrismaService) {}

  async create(dto: CreateServerDto) {
    try {
      return await this.prisma.server.create({
        data: dto
      })
    } catch (e) {
      if (e.code === 'P2002') {
        throw new ConflictException(e.meta.driverAdapterError.cause.originalMessage)
      }
      throw e
    }
  }

  async like(
    id: number,
    user: User
  ) {
    const userWithLikes = await this.prisma.user.findUnique({
      where: {
        id: user.id
      },
      select: {
        likedServers: {
          select: {
            id: true
          }
        }
      }
    })

    if (!userWithLikes) throw new UnauthorizedException('User not found')

    const isLiked = userWithLikes.likedServers.some(userLikes =>
      userLikes.id === id
    )

    try {
      await this.prisma.server.update({
        where: {
          id
        },
        data: {
          likedUsers: isLiked
            ? { disconnect: { id: user.id } }
            : { connect: { id: user.id } },
          likes: isLiked
            ? { decrement: 1 }
            : { increment: 1 }
        }
      })
    } catch (e) {
      if (e.code === 'P2025') {
        throw new BadRequestException('Server not found')
      }
      throw e
    }
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
