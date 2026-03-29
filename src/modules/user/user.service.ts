import {
  BadRequestException,
  ConflictException,
  Injectable,
  NotFoundException,
  PreconditionFailedException
} from '@nestjs/common'
import { PrismaService } from '../../prisma/prisma.service'
import { DiscordRegisterDto, RegisterDto } from '../auth/auth.dto'
import { password } from 'bun'
import { RegionDto, UpdateUserDto } from './user.dto'
import { User } from '../../../generated/prisma/client'
import {
  DeleteObjectCommand,
  PutObjectCommand,
  S3Client
} from '@aws-sdk/client-s3'
import { ConfigService } from '@nestjs/config'
import sharp from 'sharp'
import { FailedDependencyException } from '../../exceptions/failed-dependency.exception'

@Injectable()
export class UserService {
  private readonly s3Client: S3Client

  private readonly AWS_ENDPOINT: string
  private readonly AWS_BUCKET: string

  constructor(
    private readonly prisma: PrismaService,
    private readonly configService: ConfigService
  ) {
    this.AWS_ENDPOINT = configService.getOrThrow('AWS_ENDPOINT')
    this.AWS_BUCKET = configService.getOrThrow('AWS_BUCKET')

    this.s3Client = new S3Client({
      region: this.configService.getOrThrow('AWS_S3_REGION'),
      endpoint: this.AWS_ENDPOINT
    })
  }

  async getProcessedBuffer(file: Buffer, region: RegionDto) {
    const shape = sharp(file)

    const { width: imgW, height: imgH } = await shape.metadata()

    if (region.left + region.width > imgW) {
      throw new BadRequestException('Invalid region: width out of bounds')
    }

    if (region.top + region.height > imgH) {
      throw new BadRequestException('Invalid region: height out of bounds')
    }

    return await shape
      .extract(region)
      .resize({
        height: 300,
        width: 300
      })
      .webp()
      .toBuffer()
  }

  async setAvatar(file: Buffer, userId: number, region: RegionDto) {
    const buffer = await this.getProcessedBuffer(file, region)

    await this.s3Client.send(
      new PutObjectCommand({
        Bucket: this.AWS_BUCKET,
        Key: `avatars/${userId}.webp`,
        Body: buffer
      })
    )

    await this.prisma.user.update({
      where: { id: userId },
      data: {
        avatar: this.buildAWSAvatarUrl(userId)
      }
    })
  }

  async getAvatar(userId: number) {
    return (await this.findWithAvatar(userId)) ?? undefined
  }

  async findWithAvatar(userId: number) {
    const user = await this.prisma.user.findUnique({
      where: { id: userId },
      select: { avatar: true }
    })

    if (!user) throw new BadRequestException('User not found')

    return user.avatar
  }

  buildAWSAvatarUrl(userId: number) {
    return `${this.AWS_ENDPOINT}/${this.AWS_BUCKET}/avatars/${userId}.webp`
  }

  async deleteAvatar(userId: number, avatarUrl: string | null) {
    if (!avatarUrl) return

    if (avatarUrl === this.buildAWSAvatarUrl(userId)) {
      try {
        await this.s3Client.send(
          new DeleteObjectCommand({
            Bucket: this.AWS_BUCKET,
            Key: `avatars/${userId}.webp`
          })
        )
      } catch {
        throw new FailedDependencyException('Could not delete avatar from S3')
      }
    }

    await this.prisma.user.update({
      where: { id: userId },
      data: { avatar: null }
    })
  }

  async discordRegister(dto: DiscordRegisterDto) {
    return this.prisma.user.create({
      data: dto
    })
  }

  async register(dto: RegisterDto) {
    const user = await this.prisma.user.findUnique({
      where: {
        email: dto.email
      }
    })

    if (user)
      throw new ConflictException(
        `User with email ${dto.email} already exists!`
      )

    return this.prisma.user.create({
      data: {
        ...{ ...dto, iat: undefined, exp: undefined },
        password: await password.hash(dto.password)
      }
    })
  }

  async findOneOrNull(params: {
    id?: number
    email?: string
    discordId?: string
  }) {
    if (!params.id && !params.email && !params.discordId) return null

    const checkedParams = {
      id: params.id,
      email: params.email,
      discordId: params.discordId
    }

    return this.prisma.user.findUnique({
      where: checkedParams
    })
  }

  async findOneById(id: number) {
    const user = await this.prisma.user.findUnique({
      where: { id }
    })

    if (!user) throw new NotFoundException('User not found!')

    return user
  }

  async findOneForLogin(email: string) {
    const user = await this.prisma.user.findUnique({
      where: {
        email
      },
      select: {
        id: true,
        password: true
      }
    })

    if (!user) throw new NotFoundException('User not found!')

    return user
  }

  async findOneForRefresh(id: number) {
    const user = await this.prisma.user.findUnique({
      where: {
        id
      },
      select: {
        id: true
      }
    })

    if (!user) throw new NotFoundException('User not found!')

    return user
  }

  async update(id: number, dto: UpdateUserDto) {
    return this.prisma.user.update({
      where: { id },
      data: dto
    })
  }

  getUser(user: User) {
    const { email, avatar, name } = user
    return {
      name,
      email,
      avatar
    }
  }
}
