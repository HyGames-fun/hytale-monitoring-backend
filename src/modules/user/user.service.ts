import {
  ConflictException,
  Injectable,
  NotFoundException,
} from '@nestjs/common'
import { PrismaService } from '../../prisma/prisma.service'
import { DiscordRegisterDto, RegisterDto } from '../auth/auth.dto'
import { password } from 'bun'
import { UpdateUserDto } from './user.dto'

@Injectable()
export class UserService {
  constructor(private readonly prisma: PrismaService) {}

  async discordRegister(dto: DiscordRegisterDto) {
    return this.prisma.user.create({
      data: dto,
    })
  }

  async register(dto: RegisterDto) {
    const user = await this.prisma.user.findUnique({
      where: {
        email: dto.email,
      },
    })

    if (user)
      throw new ConflictException(
        `User with email ${dto.email} already exists!`,
      )

    return this.prisma.user.create({
      data: {
        ...dto,
        password: await password.hash(dto.password),
      },
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
      discordId: params.discordId,
    }

    return this.prisma.user.findUnique({
      where: checkedParams,
    })
  }

  async findOneById(id: number) {
    const user = await this.prisma.user.findUnique({
      where: { id },
    })

    if (!user) throw new NotFoundException('User not found!')

    return user
  }

  async findOneForLogin(email: string) {
    const user = await this.prisma.user.findUnique({
      where: {
        email,
      },
      select: {
        id: true,
        password: true,
      },
    })

    if (!user) throw new NotFoundException('User not found!')

    return user
  }

  async findOneForRefresh(id: number) {
    const user = await this.prisma.user.findUnique({
      where: {
        id,
      },
      select: {
        id: true,
      },
    })

    if (!user) throw new NotFoundException('User not found!')

    return user
  }

  async update(id: number, dto: UpdateUserDto) {
    return this.prisma.user.update({
      where: { id },
      data: dto,
    })
  }
}
