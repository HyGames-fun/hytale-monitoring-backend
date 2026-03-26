import { Module } from '@nestjs/common'
import { ServerService } from './server.service'
import { ServerController } from './server.controller'
import { TurnstileModule } from '../turnstile/turnstile.module'
import { SearchModule } from '../search/search.module'

@Module({
  imports: [TurnstileModule, SearchModule],
  controllers: [ServerController],
  providers: [ServerService]
})
export class ServerModule {}
