import { Controller, Get } from '@nestjs/common'
import { UserService } from './user.service'
import { Authorized } from '../../decorators/autrorized.decorator'
import type { User } from '../../../generated/prisma/client'
import { ApiBearerAuth } from '@nestjs/swagger'

@Controller('user')
export class UserController {
  constructor(private readonly userService: UserService) {}

  @ApiBearerAuth()
  @Get()
  getCurrentUser(@Authorized() user: User) {
    return this.userService.getUser(user)
  }

}