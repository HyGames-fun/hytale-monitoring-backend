import { createParamDecorator, UnauthorizedException } from '@nestjs/common'
import { User } from '../../generated/prisma/client'
import { Request } from 'express'

export const Authorized = createParamDecorator((data: keyof User, context) => {
  const request: Request = context.switchToHttp().getRequest()
  const user = request.user

  if (!user) throw new UnauthorizedException('User not found')

  return data ? user[data] : user
})
