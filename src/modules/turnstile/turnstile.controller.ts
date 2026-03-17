import { Controller } from '@nestjs/common'
import { TurnstileService } from './turnstile.service'

@Controller('turnstile')
export class TurnstileController {
  constructor(private readonly turnstileService: TurnstileService) {}
}
