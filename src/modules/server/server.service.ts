import {
  BadRequestException,
  ConflictException,
  Injectable,
  UnauthorizedException
} from '@nestjs/common'
import { CreateServerDto, FindPageDto } from './server.dto'
import { PrismaService } from '../../prisma/prisma.service'
import { User } from '../../../generated/prisma/client'
import { checkPort } from '../../validators/ip.validator'

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
        throw new ConflictException(
          e.meta.driverAdapterError.cause.originalMessage
        )
      }
      throw e
    }
  }

  async like(id: number, user: User) {
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

    const isLiked = userWithLikes.likedServers.some(
      (userLikes) => userLikes.id === id
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
          likes: isLiked ? { decrement: 1 } : { increment: 1 }
        }
      })
    } catch (e) {
      if (e.code === 'P2025') {
        throw new BadRequestException('Server not found')
      }
      throw e
    }
  }

  async findPage(dto: FindPageDto, user: User | undefined) {
    const { page, quantity, filters, order } = dto

    const servers = await this.prisma.server.findMany({
      skip: quantity * page,
      take: quantity,
      where: {
        ...(filters?.region && { region: filters.region }),
        ...(filters?.tags && {
          tags: {
            hasSome: filters.tags
          }
        })
      },
      orderBy: [
        ...(order?.likes ? [{ likes: order.likes }] : []),
        ...(order?.createdAt ? [{ createdAt: order.createdAt }] : [])
      ],
      include: {
        likedUsers: {
          select: {
            id: true
          }
        }
      }
    })

    servers.map((server) => ({
      ...server,
      liked: server.likedUsers.some((item) => item.id === user?.id),
      isOnline: server.ip ? checkPort(server.ip) : undefined,
      players: 10
    }))
  }

  findOne(id: number) {
    return `This action returns a #${id} server`
  }

  async findQuantity() {
    return this.prisma.server.count()
  }

  remove(id: number) {
    return `This action removes a #${id} server`
  }
}
