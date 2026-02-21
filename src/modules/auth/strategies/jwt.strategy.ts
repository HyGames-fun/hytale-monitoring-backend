import { Injectable } from '@nestjs/common'
import { PassportStrategy } from '@nestjs/passport'
import { ExtractJwt, Strategy } from 'passport-jwt'
import { ConfigService } from '@nestjs/config'
import { Request } from 'express'
import { Payload } from '../interfaces/payload.interface'
import { AuthService } from '../auth.service'

@Injectable()
export class JwtStrategy extends PassportStrategy(Strategy) {
  constructor(
    private readonly configService: ConfigService,
    private readonly authService: AuthService,
  ) {
    const jwtFromRequest = (req: Request): string | null =>
      ExtractJwt.fromAuthHeaderAsBearerToken()(req)

    const secretOrKey = configService.getOrThrow<string>('JWT_SECRET')

    super({
      secretOrKey,
      jwtFromRequest,
      ignoreExpiration: false,
      algorithms: ['HS256'],
    })
  }

  async validate(payload: Payload) {
    return this.authService.validate(payload)
  }
}
