import {
  ConflictException,
  Injectable,
  NotFoundException,
} from '@nestjs/common'
import { PrismaService } from '../../prisma/prisma.service'
import { RegisterDto } from '../auth/auth.dto'
import { password } from 'bun'

@Injectable()
export class UserService {
  constructor(private readonly prisma: PrismaService) {}

  async create(dto: RegisterDto) {
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

  async findOne(id?: number, email?: string) {
    if (!id && !email) return null

    const params = { id, email }

    const user = await this.prisma.user.findUnique({
      where: params,
    })

    if (!user) throw new NotFoundException('User not found!')

    return user
  }

  async findOneOrNull(id?: number, email?: string) {
    if (!id && !email) return null

    const params = { id, email }

    return this.prisma.user.findUnique({
      where: params,
    })
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
}
