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
import { Request } from 'express'
import { TurnstileService } from '../turnstile/turnstile.service'

@Injectable()
export class ServerService {
  constructor(
    private readonly prisma: PrismaService,
    private readonly turnstileService: TurnstileService
  ) {}

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

  async like(
    id: number,
    user: User | undefined,
    token: string | undefined,
    req: Request
  ) {
    if (!user) {
      await this.validateGuest(req, token)
      return this.likeAsGuest(id)
    }

    const userWithLikes = await this.getUserLikes(user.id)
    return this.toggleLike(id, user.id, userWithLikes)
  }

  private async validateGuest(req: Request, token?: string) {
    if (!token) {
      throw new BadRequestException('Invalid turnstile token!')
    }

    const isValid = await this.turnstileService.verify(req, token)

    if (!isValid) {
      throw new BadRequestException('Invalid turnstile token!')
    }
  }

  private async getUserLikes(userId: number) {
    const user = await this.prisma.user.findUnique({
      where: { id: userId },
      select: {
        likedServers: {
          select: { id: true }
        }
      }
    })

    if (!user) {
      throw new UnauthorizedException('User not found')
    }

    return user
  }

  private async toggleLike(
    serverId: number,
    userId: number,
    user: { likedServers: { id: number }[] }
  ) {
    const isLiked = user.likedServers.some((s) => s.id === serverId)

    try {
      await this.prisma.server.update({
        where: { id: serverId },
        data: {
          likedUsers: isLiked
            ? { disconnect: { id: userId } }
            : { connect: { id: userId } },
          likes: isLiked ? { decrement: 1 } : { increment: 1 }
        }
      })
    } catch (e: any) {
      if (e.code === 'P2025') {
        throw new BadRequestException('Server not found')
      }
      throw e
    }
  }

  private async likeAsGuest(serverId: number) {
    try {
      await this.prisma.server.update({
        where: { id: serverId },
        data: {
          likes: { increment: 1 }
        }
      })
    } catch (e: any) {
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
