import { Injectable } from '@nestjs/common'
import { HttpService } from '@nestjs/axios'
import { Request } from 'express'
import { ConfigService } from '@nestjs/config'
import { firstValueFrom } from 'rxjs'

@Injectable()
export class TurnstileService {
  private readonly SECRET_KEY: string
  constructor(
    private readonly httpService: HttpService,
    private readonly configService: ConfigService
  ) {
    this.SECRET_KEY = configService.getOrThrow('TURNSTILE_SECRET_KEY')
  }

  async verify(req: Request, token: string) {
    const xForwardedFor = req.headers['x-forwarded-for']
    const forwardedIp = Array.isArray(xForwardedFor)
      ? xForwardedFor[0]
      : xForwardedFor?.split(',')[0]?.trim()

    const ip = forwardedIp || req.ip || 'unknown'

    try {
      const response = await firstValueFrom(
        this.httpService.post(
          'https://challenges.cloudflare.com/turnstile/v0/siteverify',
          {
            secret: this.SECRET_KEY,
            response: token,
            remoteip: ip
          },
          {
            headers: {
              'Content-Type': 'application/json'
            }
          }
        )
      )
      return response.data.success as boolean
    } catch {
      return false
    }
  }
}
