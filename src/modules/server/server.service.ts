import {
  BadRequestException,
  ConflictException,
  Inject,
  Injectable,
  NotFoundException,
  UnauthorizedException
} from '@nestjs/common'
import {
  CreateServerDto,
  FindPageDto,
  ServerDto,
  UpdateServerDto
} from './server.dto'
import { PrismaService } from '../../prisma/prisma.service'
import { Tag, User } from '../../../generated/prisma/client'
import { checkPort } from '../../validators/ip.validator'
import { Request } from 'express'
import { TurnstileService } from '../turnstile/turnstile.service'
import { CACHE_MANAGER, Cache } from '@nestjs/cache-manager'
import ms from 'ms'
import { SearchService } from '../search/search.service'
import { ConfigService } from '@nestjs/config'

@Injectable()
export class ServerService {
  private readonly SEARCH_METHOD: string

  constructor(
    private readonly prisma: PrismaService,
    private readonly turnstileService: TurnstileService,
    private readonly searchService: SearchService,
    private readonly configService: ConfigService,
    @Inject(CACHE_MANAGER) private cacheManager: Cache
  ) {
    this.SEARCH_METHOD = configService.getOrThrow('SEARCH_METHOD')
  }

  async create(dto: CreateServerDto) {
    let server: ServerDto
    try {
      server = await this.prisma.server.create({
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

    if (this.SEARCH_METHOD === 'elasticsearch') {
      await this.searchService.indexServer(server)
    }

    return server
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
      const isLiked = await this.guestLikeCheck(req, id, true)
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

    return {}
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

  private async guestLikeCheck(req: Request, id: number, change: boolean) {
    const ip = req.ip || undefined

    if (!ip) throw new BadRequestException('Guest IP not found')

    const isLiked: boolean | undefined = await this.cacheManager.get(
      `guest_like:${ip}:${id}`
    )

    if (change) {
      if (!isLiked) {
        await this.cacheManager.set(`guest_like:${ip}:${id}`, true, ms('1d'))
      } else {
        await this.cacheManager.del(`guest_like:${ip}:${id}`)
      }
    }

    return isLiked ?? false
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

  private async likeAsGuest(serverId: number, isLiked?: boolean) {
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
        likes: user
          ? {
              where: { userId: user.id },
              select: { serverId: true }
            }
          : false
      },
      omit: {
        region: true,
        userId: true,
        createdAt: true,
        updatedAt: true
      }
    })

    return Promise.all(
      servers.map(async (server) => {
        let liked: boolean

        if (user) {
          liked = server.likes?.length > 0
        } else {
          liked = await this.guestLikeCheck(req, server.id, false)
        }

        return {
          ...server,
          likes: undefined,
          liked,
          players: 10
        }
      })
    )
  }

  async getOnlineStatus(serverId: number) {
    const server = await this.prisma.server.findUnique({
      where: {
        id: serverId
      },
      select: {
        id: true,
        ip: true
      }
    })

    if (!server) throw new BadRequestException('Server not found')

    return {
      id: server.id,
      isOnline: server.ip ? await checkPort(server.ip) : null
    }
  }

  async getOnlineStatuses(serverIds: number[]) {
    const servers = await this.prisma.server.findMany({
      where: {
        id: { in: serverIds }
      },
      select: {
        id: true,
        ip: true
      }
    })

    if (!servers.length) throw new BadRequestException('Servers not found')

    return Promise.all(
      servers.map(async (server) => ({
        id: server.id,
        isOnline: server.ip ? await checkPort(server.ip) : null
      }))
    )
  }

  async findOne(nameId: string, user: User | undefined, req: Request) {
    const server = await this.prisma.server.findUnique({
      where: { nameId },
      include: {
        likes: user
          ? {
              where: { userId: user.id },
              select: { serverId: true }
            }
          : false
      },
      omit: {
        userId: true,
        createdAt: true,
        updatedAt: true
      }
    })

    console.log(server)

    if (!server) throw new BadRequestException('Server not found')

    let liked: boolean

    if (user) {
      liked = server.likes?.length > 0
    } else {
      liked = await this.guestLikeCheck(req, server.id, false)
    }

    return {
      ...server,
      likes: undefined,
      liked,
      players: 10
    }
  }

  async findQuantity() {
    return this.prisma.server.count()
  }

  async search(q: string, page: number, limit: number) {
    if (!q) throw new BadRequestException('Query is undefined')

    const normalized = q.trim().toLowerCase()

    if (!normalized) throw new BadRequestException('Query is empty')

    if (limit < 1) throw new BadRequestException('Limit is less than 1')
    if (limit > 50) throw new BadRequestException('Limit is more than 50')

    if (page < 0) throw new BadRequestException('Limit is less than 0')

    if (this.SEARCH_METHOD === 'elasticsearch') {
      const search = await this.searchService.searchServers(q, page, limit)
      const ids = search.map((s) => s.id)

      return this.prisma.server.findMany({
        where: {
          id: { in: ids }
        }
      })
    }

    const TAG_VALUES = Object.values(Tag)

    const words = [...new Set(normalized.split(/\s+/).filter(Boolean))]

    if (words.length === 0) {
      throw new BadRequestException('Query is empty')
    }

    return this.prisma.server.findMany({
      skip: page * limit,
      take: limit,
      where: {
        AND: words.map((word) => {
          const tagCandidate = word.toUpperCase()

          const isTag = TAG_VALUES.includes(tagCandidate as Tag)

          return {
            OR: [
              {
                name: {
                  contains: word,
                  mode: 'insensitive'
                }
              },
              {
                description: {
                  contains: word,
                  mode: 'insensitive'
                }
              },
              ...(isTag
                ? [
                    {
                      tags: {
                        has: tagCandidate as Tag
                      }
                    }
                  ]
                : [])
            ]
          }
        })
      }
    })
  }

  async update(id: number, dto: UpdateServerDto) {
    let server: ServerDto
    try {
      server = await this.prisma.server.update({
        where: { id },
        data: dto
      })
    } catch (e) {
      if (e.code === 'P2025') {
        throw new NotFoundException('Server not found')
      }
      throw e
    }

    if (this.SEARCH_METHOD === 'elasticsearch') {
      await this.searchService.indexServer(server)
    }

    return server
  }

  async remove(id: number) {
    let server: ServerDto
    try {
      server = await this.prisma.server.delete({
        where: { id }
      })
    } catch (e) {
      if (e.code === 'P2025') {
        throw new NotFoundException('Server not found')
      }
      throw e
    }

    if (this.SEARCH_METHOD === 'elasticsearch') {
      await this.searchService.deleteServer(id)
    }

    return server
  }
}
