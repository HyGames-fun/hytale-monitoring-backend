import {
  BadRequestException,
  ConflictException,
  Inject,
  Injectable,
  UnauthorizedException
} from '@nestjs/common'
import { CreateServerDto, FindPageDto } from './server.dto'
import { PrismaService } from '../../prisma/prisma.service'
import { User } from '../../../generated/prisma/client'
import { checkPort } from '../../validators/ip.validator'
import { Request } from 'express'
import { TurnstileService } from '../turnstile/turnstile.service'
import { CACHE_MANAGER, Cache } from '@nestjs/cache-manager'
import ms from 'ms'

@Injectable()
export class ServerService {
  constructor(
    private readonly prisma: PrismaService,
    private readonly turnstileService: TurnstileService,
    @Inject(CACHE_MANAGER) private cacheManager: Cache
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
    await this.checkServerId(id)

    if (!user) {
      await this.validateGuest(req, token)
      const isLiked = await this.guestLikeCheck(req, id)
      return this.likeAsGuest(id, isLiked)
    }

    const userWithLikes = await this.getUserLikes(user.id)
    return this.toggleLike(id, user.id, userWithLikes)
  }

  private async checkServerId(id: number) {
    const server = await this.prisma.server.findUnique({
      where: {
        id
      },
      select: {
        id: true
      }
    })

    if (!server) throw new BadRequestException('Server not found')
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

  private async guestLikeCheck(req: Request, id: number) {
    const ip = req.ip || undefined

    if (!ip) throw new BadRequestException('Guest IP not found')

    const isLiked = await this.cacheManager.get(`guest_like:${ip}:${id}`)

    if (isLiked) {
      await this.cacheManager.del(`guest_like:${ip}:${id}`)
      return true
    }

    await this.cacheManager.set(`guest_like:${ip}:${id}`, true, ms('1d'))
    return false
  }

  private async getUserLikes(userId: number) {
    const user = await this.prisma.user.findUnique({
      where: { id: userId },
      select: {
        likes: {
          select: { serverId: true }
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
    user: { likes: { serverId: number }[] }
  ) {
    const isLiked = user.likes.some((s) => s.serverId === serverId)

    await this.prisma.$transaction([
      isLiked
        ? this.prisma.like.delete({
            where: { userId_serverId: { userId, serverId } }
          })
        : this.prisma.like.create({ data: { userId, serverId } }),
      this.prisma.server.update({
        where: { id: serverId },
        data: { likesQuantity: isLiked ? { decrement: 1 } : { increment: 1 } }
      })
    ])
  }

  private async likeAsGuest(serverId: number, isLiked: boolean) {
    await this.prisma.server.update({
      where: { id: serverId },
      data: {
        likesQuantity: isLiked ? { decrement: 1 } : { increment: 1 }
      }
    })
  }

  async findPage(dto: FindPageDto, req: Request, user?: User) {
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
        ...(order?.likes ? [{ likesQuantity: order.likes }] : []),
        ...(order?.createdAt ? [{ createdAt: order.createdAt }] : [])
      ],
      include: {
        likes: {
          select: {
            userId: true
          }
        }
      }
    })

    servers.map((server) => ({
      ...server,
      liked: user
        ? server.likes.some((item) => item.userId === user.id)
        : this.guestLikeCheck(req, server.id),
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
